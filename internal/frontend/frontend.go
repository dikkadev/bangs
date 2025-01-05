package frontend

import (
	"bangs/web"
	"net/http"

	"github.com/a-h/templ"
)

func Handler() http.Handler {
	router := http.NewServeMux()

	router.Handle("/", templ.Handler(web.Index()))

	return router
}
