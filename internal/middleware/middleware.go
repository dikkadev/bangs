package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func CreateStack(wares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(wares) - 1; i >= 0; i-- {
			next = wares[i](next)
		}
		return next
	}

}
