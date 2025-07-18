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
	Default QueryURL          `yaml:"default" json:"default"`
	Aliases map[string]string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	Entries BangList          `yaml:",inline" json:"bangs"`
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

		if len(registry.Aliases) > 0 {
			aliasKeys := make([]string, 0, len(registry.Aliases))
			for k := range registry.Aliases {
				aliasKeys = append(aliasKeys, k)
			}
			slog.Debug("All loaded aliases", "aliases", registry.Aliases)
		}
	}
	return nil
}

func (r *Registry) DefaultForward(query string, w http.ResponseWriter, req *http.Request) error {
	defaultStr := string(r.Default)

	// Check if default is a bang reference (doesn't contain ://)
	if !strings.Contains(defaultStr, "://") {
		// Handle as bang reference(s)
		return r.handleDefaultBangReferences(defaultStr, query, w, req)
	}

	// Handle as traditional URL
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

func (r *Registry) handleDefaultBangReferences(defaultStr, query string, w http.ResponseWriter, req *http.Request) error {
	// Parse bang references (support multi-bang with +)
	bangRefs := strings.Split(defaultStr, "+")
	entries := make([]*Entry, 0, len(bangRefs))

	// Resolve each bang reference to an Entry
	for _, bangRef := range bangRefs {
		bangRef = strings.TrimSpace(bangRef)
		if bangRef == "" {
			continue
		}

		// Check if it's an alias first
		if alias, exists := r.Aliases[bangRef]; exists {
			slog.Debug("Resolved alias in default", "alias", bangRef, "target", alias)
			// Recursively handle the alias target (which might be multi-bang)
			aliasRefs := strings.Split(alias, "+")
			for _, aliasRef := range aliasRefs {
				aliasRef = strings.TrimSpace(aliasRef)
				if aliasRef == "" {
					continue
				}
				entry, exists := r.Entries.byBang[aliasRef]
				if !exists {
					slog.Error("Default alias target bang not found", "alias", bangRef, "target", aliasRef)
					http.Error(w, fmt.Sprintf("Default alias '%s' target bang '%s' not found", bangRef, aliasRef), http.StatusInternalServerError)
					return fmt.Errorf("default alias '%s' target bang '%s' not found", bangRef, aliasRef)
				}
				entries = append(entries, &entry)
			}
		} else {
			// Regular bang lookup
			entry, exists := r.Entries.byBang[bangRef]
			if !exists {
				slog.Error("Default bang reference not found", "bang", bangRef)
				http.Error(w, fmt.Sprintf("Default bang reference '%s' not found", bangRef), http.StatusInternalServerError)
				return fmt.Errorf("default bang reference '%s' not found", bangRef)
			}
			entries = append(entries, &entry)
		}
	}

	if len(entries) == 0 {
		http.Error(w, "No valid bang references found in default", http.StatusInternalServerError)
		return fmt.Errorf("no valid bang references found in default")
	}

	slog.Debug("Default bang resolution complete", "entryCount", len(entries), "entries", func() []string {
		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Bang
		}
		return names
	}())

	// Handle single vs multi-bang
	if len(entries) == 1 {
		// Single bang - direct redirect
		slog.Debug("Single bang default, redirecting", "bang", entries[0].Bang)
		return entries[0].Forward(query, w, req)
	}

	slog.Debug("Multi-bang default, generating HTML", "bangCount", len(entries))
	return generateMultiTabHTML(entries, query, w)
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
	var tempMap map[string]any
	err := value.Decode(&tempMap)
	if err != nil {
		return err
	}
	delete(tempMap, "default")
	delete(tempMap, "aliases")

	bl.Entries = make(map[string]Entry, len(tempMap))
	bl.byBang = make(map[string]Entry, len(tempMap))
	for k, a := range tempMap {
		// Skip non-map entries (like aliases which are strings)
		v, ok := a.(map[string]any)
		if !ok {
			continue
		}
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

	// Resolve alias if it exists
	if alias, exists := registry.Aliases[rawBang]; exists {
		slog.Debug("Resolved alias", "alias", rawBang, "target", alias)
		rawBang = alias
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
