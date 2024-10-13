package middleware

import (
	"net/http"
	"strings"
)

func BlockPathEndingInSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.Error(w, "405 not allowed :(", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}
