package bangs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"bangs/pkg/middleware"

	"github.com/dikkadev/prettyslog"
)

func Handler(doAllowNoBang bool, ignoreCharPar string) http.Handler {
	allowNoBang = doAllowNoBang
	ignoreChar = ignoreCharPar

	router := http.NewServeMux()

	router.HandleFunc("/list", listAll)
	router.HandleFunc("/", searchByQuery)
	router.HandleFunc("/{bang}/{query...}", searchByPath)

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
		slog.Error("Error converting registry to json", "err", err)
		http.Error(w, fmt.Sprintf("Internal JSON error -.-\n%v\n", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
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

	if strings.HasPrefix(q, "##") {
		slog.Debug("Double hashtag found, forwarding to default", "query", q)
		q = strings.TrimPrefix(q, "##")
		_ = registry.DefaultForward(q, w, r)
		return
	}

	entry, query, err := registry.Entries.PrepareInput(q)
	if err != nil {
		if _, ok := err.(InputHasNoBangError); ok {
			slog.Debug("No bang found in input, forwarding to default", "query", q)
			_ = registry.DefaultForward(q, w, r)
			return
		}
		if _, ok := err.(InputStartsWithIgnoreError); ok {
			slog.Debug("Input starts with ignore character, removing ingoreChar and forwarding to default", "query", q)
			q = q[1:]
			_ = registry.DefaultForward(q, w, r)
			return
		}

		slog.Error("Error preparing input", "err", err)
		http.Error(w, fmt.Sprintf("Error preparing input: %v", err), http.StatusBadRequest)
		return
	}

	_ = entry.Forward(query, w, r)
}

func searchByPath(w http.ResponseWriter, r *http.Request) {
	// Extract bang and query from path segments
	bang := r.PathValue("bang")
	if bang == "" {
		msg := "No bang provided in path"
		slog.Error(msg, "url", r.URL)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// PathValue automatically decodes the path segment
	query := r.PathValue("query")
	// query can be empty if the path is just /bang/

	entry, ok := registry.Entries.byBang[bang]
	if !ok {
		// If bang is not found, try default forward if allowed?
		// Current logic: return 404
		msg := fmt.Sprintf("Unknown bang: '%s'", bang)
		slog.Debug(msg, "url", r.URL)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	slog.Debug("Forwarding search by path", "bang", bang, "query", query)
	_ = entry.Forward(query, w, r)
}
