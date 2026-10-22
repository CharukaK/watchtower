package app

import (
	"net/http"

	"github.com/CharukaK/watchtower/internal/dashboard"
	"github.com/CharukaK/watchtower/web"
	"github.com/a-h/templ"
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServer(http.FS(web.Static)))
	mux.Handle("GET /{$}", templ.Handler(dashboard.Dashboard()))
	return mux
}
