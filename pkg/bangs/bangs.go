package bangs

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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
	Bang        string   `yaml:"bang" json:"bang"`
	Description string   `yaml:"description" json:"description"`
	URL         QueryURL `yaml:"url" json:"url"`
	Category    string   `yaml:"category,omitempty" json:"category,omitempty"`
}

func (e Entry) String() string {
	if e.Description == "" {
		return fmt.Sprintf("{bang: %s, url: %s}", e.Bang, e.URL)
	}
	return fmt.Sprintf("{bang: %s, description: %s, url: %s}", e.Bang, e.Description, e.URL)
}

func (e Entry) Equals(other Entry) bool {
	return e.Bang == other.Bang && e.Description == other.Description && e.URL == other.URL && e.Category == other.Category
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

func generateMultiTabHTML(entries []*Entry, query string, w http.ResponseWriter) error {
	urls := make([]string, len(entries))
	for i, entry := range entries {
		u, err := entry.URL.Augment(query)
		if err != nil {
			slog.Error("Error augmenting URL", "err", err)
			if _, ok := err.(AugmentNoPlaceholderError); ok {
				http.Error(w, "No placeholder found in path, query, or fragment", http.StatusBadRequest)
			} else {
				http.Error(w, fmt.Sprintf("Error augmenting URL: %v", err), http.StatusInternalServerError)
			}
			return err
		}
		urls[i] = u.String()
	}

	var newTabCommands []string
	for i := 1; i < len(urls); i++ {
		newTabCommands = append(newTabCommands, fmt.Sprintf("window.open('%s', '_blank');", urls[i]))
	}
	newTabsJS := strings.Join(newTabCommands, "\n")

	w.Header().Set("Content-Type", "text/html")
	hypertext := fmt.Sprintf(`<html>
<head></head>
<body>
<script>
%s
window.location.href = "%s";
</script>
</body>
</html>`, newTabsJS, urls[0])

	w.Write([]byte(hypertext))
	return nil
}
