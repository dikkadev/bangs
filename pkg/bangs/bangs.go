package bangs

import (
	"bangs/internal/middleware"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sett17/prettyslog"
	"gopkg.in/yaml.v3"
)

type QueryURL string

type AugmentNoPlaceholderError struct{}

func (AugmentNoPlaceholderError) Error() string {
	return "no placeholder found in path, query, or fragment"
}

func (q QueryURL) Augment(query string) (*url.URL, error) {
	const placeholder = "{}"
	queryURLStr := string(q)

	u, err := url.Parse(queryURLStr)
	if err != nil {
		return nil, err
	}

	placeholderFound := false

	if strings.Contains(u.Path, placeholder) {
		u.Path = strings.ReplaceAll(u.Path, placeholder, url.PathEscape(query))
		placeholderFound = true
	}

	if strings.Contains(u.RawQuery, placeholder) {
		u.RawQuery = strings.ReplaceAll(u.RawQuery, placeholder, url.QueryEscape(query))
		placeholderFound = true
	}

	if strings.Contains(u.Fragment, placeholder) {
		u.Fragment = strings.ReplaceAll(u.Fragment, placeholder, url.QueryEscape(query))
		placeholderFound = true
	}

	if !placeholderFound {
		return nil, AugmentNoPlaceholderError{}
	}

	return u, nil
}

type Entry struct {
	Bang string   `yaml:"bang" json:"bang"`
	URL  QueryURL `yaml:"url" json:"url"`
}

func (e Entry) Forward(query string, w http.ResponseWriter, r *http.Request) error {
	u, err := e.URL.Augment(query)
	if err != nil {
		slog.Error("Error augmenting URL", "err", err)
		if _, ok := err.(AugmentNoPlaceholderError); ok {
			http.Error(w, "No placeholder found in path, query, or fragment", http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Error augmenting URL: %v", err), http.StatusInternalServerError)
		}
		return err
	}

	http.Redirect(w, r, u.String(), http.StatusFound)
	return nil
}

type BangList struct {
	Entries map[string]Entry
	byBang  map[string]Entry
	len     int
}

func (bl *BangList) UnmarshalYAML(value *yaml.Node) error {
	var thisisstupid map[string]any
	err := value.Decode(&thisisstupid)
	if err != nil {
		return err
	}
	delete(thisisstupid, "default")

	bl.Entries = make(map[string]Entry, len(thisisstupid))
	bl.byBang = make(map[string]Entry, len(thisisstupid))
	for k, a := range thisisstupid {
		v := a.(map[string]any)
		bangChars, ok := v["bang"].(string)
		if !ok {
			return fmt.Errorf("missing bang field for entry '%s'", k)
		}
		bangChars = strings.TrimSpace(bangChars)
		if len(bangChars) == 0 {
			return fmt.Errorf("bang field is empty for entry '%s'", k)
		}
		if _, ok := bl.byBang[bangChars]; ok {
			return fmt.Errorf("duplicate bang found for '%s': %s", k, bangChars)
		}
		url, ok := v["url"].(string)
		if !ok {
			return fmt.Errorf("missing url field for entry %s", k)
		}
		entry := Entry{
			Bang: bangChars,
			URL:  QueryURL(url),
		}
		bl.Entries[k] = entry
		bl.byBang[bangChars] = entry
	}
	bl.len = len(thisisstupid)

	return nil
}

type InputHasNoBangError string

func (InputHasNoBangError) Error() string {
	return "input does not contain a bang"
}

func (bl BangList) PrepareInput(input string) (*Entry, string, error) {
	if len(input) < 2 {
		return nil, "", fmt.Errorf("len(query) was smaller than, which is not valid")
	}
	if input[0] != '!' {
		return nil, input, InputHasNoBangError(input)
	}
	split := strings.SplitN(input[1:], " ", 2) // should this also allow for tabs?
	if len(split) != 2 {
		return nil, "", fmt.Errorf("query does not contain a bang and a query")
	}

	bang, query := split[0], split[1]

	entry, ok := bl.byBang[bang]
	if !ok {
		return nil, "", fmt.Errorf("unknown bang: '%s'", bang)
	}
	return &entry, query, nil
}

func (bl BangList) PrepareInputNaive(input string) (*Entry, string, error) {
	if len(input) < 2 {
		return nil, "", fmt.Errorf("len(query) was smaller than, which is not valid")
	}
	if input[0] != '!' {
		return nil, "", fmt.Errorf("query does not start with a '!'")
	}
	split := strings.SplitN(input[1:], " ", 2) // should this also allow for tabs?
	if len(split) != 2 {
		return nil, "", fmt.Errorf("query does not contain a bang and a query")
	}

	bang, query := split[0], split[1]

	for _, entry := range bl.Entries {
		if entry.Bang == bang {
			return &entry, query, nil
		}
	}
	return nil, "", fmt.Errorf("unknown bang: '%s'", bang)
}

type Config struct {
	Default QueryURL `yaml:"default" json:"default"`
	Entries BangList `yaml:",inline" json:"bangs"`
}

func (c *Config) DefaultForward(query string, w http.ResponseWriter, r *http.Request) error {
	u, err := c.Default.Augment(query)
	if err != nil {
		slog.Error("Error augmenting URL", "err", err)
		if _, ok := err.(AugmentNoPlaceholderError); ok {
			http.Error(w, "No placeholder found in path, query, or fragment", http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Error augmenting URL: %v", err), http.StatusInternalServerError)
		}
		return err
	}

	http.Redirect(w, r, u.String(), http.StatusFound)
	return nil
}

var config *Config

func All() BangList {
	return config.Entries
}

func Load(path string) error {
	slog.Info("Loading bang configuration", "file", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	config = &cfg
	slog.Info("Loaded bang configuration", "N", len(config.Entries.Entries))
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		keys := make([]string, 0, len(config.Entries.Entries))
		for k := range config.Entries.Entries {
			keys = append(keys, k)
		}
		slog.Debug("All loaded bangs", "names", keys)
	}
	return nil
}

func Handler() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /list", listAll)
	router.HandleFunc("GET /", searchByQuery)
	router.HandleFunc("GET /{bang}/{query...}", searchByPath)

	logOptions := make([]prettyslog.Option, 0)
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		logOptions = append(logOptions, prettyslog.WithLevel(slog.LevelDebug))
	}
	logger := slog.New(prettyslog.NewPrettyslogHandler("HTTP", logOptions...))
	stack := middleware.CreateStack(
		middleware.Logger(logger, "bang"),
	)

	return stack(router)
}

func listAll(w http.ResponseWriter, r *http.Request) {
	asJSON, err := json.Marshal(All().Entries)
	if err != nil {
		slog.Error("Error converting config to json", "err", err)
		http.Error(w, fmt.Sprintf("Internal JSON error -.-\n%v\n", err), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(asJSON)
	if err != nil {
		slog.Error("Error writing response", "err", err)
		return
	}
}

func searchByQuery(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	q := queries.Get("q")
	if strings.TrimSpace(q) == "" {
		msg := "No query provided for search"
		slog.Error(msg, "url", r.URL)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	entry, query, err := config.Entries.PrepareInput(q)
	if err != nil {
		if _, ok := err.(InputHasNoBangError); ok {
			slog.Debug("No bang found in input, forwarding to default", "query", q)
			_ = config.DefaultForward(q, w, r)
			return
		}
		slog.Error("Error preparing input", "err", err)
		http.Error(w, fmt.Sprintf("Error preparing input: %v", err), http.StatusBadRequest)
		return
	}

	_ = entry.Forward(query, w, r)
}

func searchByPath(w http.ResponseWriter, r *http.Request) {
	bang := r.PathValue("bang")
	bang = strings.TrimSpace(bang)
	if len(bang) == 0 {
		msg := "No bang provided for search"
		slog.Error(msg, "url", r.URL)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	query := r.PathValue("query")
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		msg := "No query provided for search"
		slog.Error(msg, "url", r.URL)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	entry, ok := config.Entries.byBang[bang]
	if !ok {
		msg := fmt.Sprintf("Unknown bang: '%s'", bang)
		slog.Error(msg, "url", r.URL)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	_ = entry.Forward(query, w, r)
}
