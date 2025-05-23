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

func Handler(doAllowNoBang bool, doAllowMultiBang bool, ignoreCharPar string) http.Handler {
	allowNoBang = doAllowNoBang
	ignoreChar = ignoreCharPar
	allowMultiBang = doAllowMultiBang

	router := http.NewServeMux()

	router.HandleFunc("/list", listAll)
	router.HandleFunc("/", searchByQuery)

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

func openMultiWithJS(w http.ResponseWriter, r *http.Request, entries []*Entry, query string) {
	if len(entries) > 2 {
		slog.Warn("More than 2 entries found, can only use the first and last")
	}

	w.Header().Set("Content-Type", "text/html")
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
			return
		}
		urls[i] = u.String()
	}

	newTab := fmt.Sprintf("window.open('%s', '_blank');\n", urls[len(urls)-1])

	hypertext := fmt.Sprintf(`
<html>
<head>
</head>
<body>
<script>
%s
window.location.href = "%s";
</script>
</body>
</html>
	`,
		newTab,
		urls[:1][0],
	)

	w.Write([]byte(hypertext))
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

	entries, query, err := registry.Entries.PrepareInput(q)
	if len(entries) > 1 {
		openMultiWithJS(w, r, entries, query)
		return
	}

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
	entry := entries[0]
	_ = entry.Forward(query, w, r)
}
