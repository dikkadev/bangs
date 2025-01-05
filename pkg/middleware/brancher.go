package middleware

import "net/http"

func BranchOnQuery(withoutQuery http.Handler, withQuery http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query()) == 0 {
			withoutQuery.ServeHTTP(w, r)
			return
		}
		withQuery.ServeHTTP(w, r)
	})
}
