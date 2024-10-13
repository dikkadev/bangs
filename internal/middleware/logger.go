package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	StatusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func Logger(logger *slog.Logger, msg string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			queries := r.URL.Query()
			writer := &responseWriterWrapper{w, http.StatusOK}
			next.ServeHTTP(writer, r)
			timeTaken := time.Since(start)

			logger.Info(msg, "status", writer.StatusCode, "method", r.Method, "path", r.URL.Path, "queries", queries, "duration", timeTaken)
		})
	}
}
