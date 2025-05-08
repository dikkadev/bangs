package bangs

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var registry *Registry

type Registry struct {
	Default QueryURL `yaml:"default" json:"default"`
	Entries BangList `yaml:",inline" json:"bangs"`
}

var allowNoBang = false
var ignoreChar = "."
var allowMultiBang = false

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var reg Registry
	err = yaml.Unmarshal(data, &reg)
	if err != nil {
		return err
	}

	debugEnabled := slog.Default().Enabled(context.Background(), slog.LevelDebug)

	if registry != nil && debugEnabled {
		diffRegistry(registry, &reg)
	}

	registry = &reg
	slog.Info("Loaded bang registry", "file", path, "N", len(registry.Entries.Entries))
	if debugEnabled {
		keys := make([]string, 0, len(registry.Entries.Entries))
		for k := range registry.Entries.Entries {
			keys = append(keys, k)
		}
		slog.Debug("All loaded bangs", "names", keys)
	}
	return nil
}

func (r *Registry) DefaultForward(query string, w http.ResponseWriter, req *http.Request) error {
	u, err := r.Default.Augment(query)
	if err != nil {
		slog.Error("Error augmenting URL", "err", err)
		if _, ok := err.(AugmentNoPlaceholderError); ok {
			http.Error(w, "No placeholder found in path, query, or fragment", http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Error augmenting URL: %v", err), http.StatusInternalServerError)
		}
		return err
	}

	http.Redirect(w, req, u.String(), http.StatusFound)
	return nil
}

func diffRegistry(oldRegistry, newRegistry *Registry) {
	if oldRegistry.Default != newRegistry.Default {
		slog.Info("Default bang URL changed", "old", oldRegistry.Default, "new", newRegistry.Default)
	}

	oldEntries := oldRegistry.Entries.Entries
	newEntries := newRegistry.Entries.Entries

	for key, newEntry := range newEntries {
		oldEntry, exists := oldEntries[key]
		if !exists {
			slog.Debug("Added bang entry", "name", key, "entry", newEntry)
		} else {
			if !oldEntry.Equals(newEntry) {
				slog.Debug("Changed bang entry", "name", key, "old", oldEntry, "new", newEntry)
			}
		}
	}

	for key, oldEntry := range oldEntries {
		if _, exists := newEntries[key]; !exists {
			slog.Debug("Removed bang entry", "name", key, "entry", oldEntry)
		}
	}
}

func All() BangList {
	return registry.Entries
}

type BangList struct {
	Entries map[string]Entry
	byBang  map[string]Entry
	len     int
}

func (bl *BangList) UnmarshalYAML(value *yaml.Node) error {
	var tempMap map[string]interface{}
	err := value.Decode(&tempMap)
	if err != nil {
		return err
	}
	delete(tempMap, "default")

	bl.Entries = make(map[string]Entry, len(tempMap))
	bl.byBang = make(map[string]Entry, len(tempMap))
	for k, a := range tempMap {
		v := a.(map[string]interface{})
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
		urlStr, ok := v["url"].(string)
		if !ok {
			return fmt.Errorf("missing url field for entry %s", k)
		}
		description, ok := v["description"].(string)
		if !ok {
			description = ""
		}
		category, ok := v["category"].(string)
		if !ok {
			category = ""
		}
		entry := Entry{
			Bang:        bangChars,
			URL:         QueryURL(urlStr),
			Description: description,
			Category:    category,
		}
		bl.Entries[k] = entry
		bl.byBang[bangChars] = entry
	}
	bl.len = len(tempMap)

	return nil
}

type InputHasNoBangError string

func (InputHasNoBangError) Error() string {
	return "input does not contain a bang"
}

type InputStartsWithIgnoreError string

func (InputStartsWithIgnoreError) Error() string {
	return "input starts with ignore character"
}

func (bl BangList) PrepareInputOld(input string) (*Entry, string, error) {
	if !allowNoBang && len(input) < 2 {
		return nil, "", fmt.Errorf("len(query) was smaller than 2, which is not valid")
	}
	if input[0] == ignoreChar[0] {
		return nil, "", InputStartsWithIgnoreError(input)
	}
	bangOffset := 1
	if input[0] != '!' {
		if !allowNoBang {
			return nil, "", InputHasNoBangError(input)
		}
		bangOffset = 0
	}

	split := strings.SplitN(input[bangOffset:], " ", 2)
	if !allowNoBang && len(split) != 2 {
		return nil, "", fmt.Errorf("query does not contain a bang and a query")
	}

	var (
		bang, query string
	)
	if len(split) == 1 {
		query = split[0]
	} else {
		bang, query = split[0], split[1]
	}

	entry, ok := bl.byBang[bang]
	if !ok {
		if allowNoBang {
			return nil, "", InputHasNoBangError(input)
		}

		return nil, "", fmt.Errorf("unknown bang: '%s'", bang)
	}
	return &entry, query, nil
}

func (bl BangList) PrepareInput(input string) ([]*Entry, string, error) {
	entries := make([]*Entry, 0)
	if !allowNoBang && len(input) < 2 {
		return nil, "", fmt.Errorf("len(query) was smaller than 2, which is not valid")
	}
	if input[0] == ignoreChar[0] {
		return nil, "", InputStartsWithIgnoreError(input)
	}
	bangOffset := 1
	if input[0] != '!' {
		if !allowNoBang {
			return nil, "", InputHasNoBangError(input)
		}
		bangOffset = 0
	}

	split := strings.SplitN(input[bangOffset:], " ", 2)
	if !allowNoBang && len(split) != 2 {
		return nil, "", fmt.Errorf("query does not contain a bang and a query")
	}

	var (
		rawBang, query string
	)
	if len(split) == 1 {
		query = split[0]
	} else {
		rawBang, query = split[0], split[1]
	}

	bangs := make([]string, 0)
	if strings.ContainsRune(rawBang, '+') {
		bangs = strings.Split(rawBang, "+")
	} else {
		bangs = append(bangs, rawBang)
	}
	slog.Debug("Parsed bangs", "bangs", bangs)

	for _, bang := range bangs {
		entry, ok := bl.byBang[bang]
		if !ok {
			if allowNoBang {
				return nil, "", InputHasNoBangError(input)
			}

			return nil, "", fmt.Errorf("unknown bang: '%s'", bang)
		}
		entries = append(entries, &entry)
	}
	return entries, query, nil
}

// Benchmarked; lookup in precomputed map is faster even in smaller cases
func (bl BangList) PrepareInputNaive(input string) (*Entry, string, error) {
	if len(input) < 2 {
		return nil, "", fmt.Errorf("len(query) was smaller than 2, which is not valid")
	}
	if input[0] != '!' {
		return nil, "", fmt.Errorf("query does not start with a '!'")
	}
	split := strings.SplitN(input[1:], " ", 2)
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

// ListAllBangs returns the map of all loaded bang entries.
func ListAllBangs() map[string]Entry {
	if registry == nil {
		return make(map[string]Entry) // Return empty map if not loaded
	}
	return registry.Entries.Entries
}
