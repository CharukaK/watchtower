package app

import (
	"log/slog"
	"net/http"

	"github.com/CharukaK/watchtower/internal/backend/isapi"
	"github.com/CharukaK/watchtower/internal/dashboard"
	"github.com/CharukaK/watchtower/web"
	"github.com/a-h/templ"
)

func New(isapiClient *isapi.Client) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServer(http.FS(web.Static)))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		info, err := isapiClient.GetDeviceInfo(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get device information", "error", err)
			http.Error(w, "device unavailable", http.StatusBadGateway)
			return
		}
		templ.Handler(dashboard.Dashboard(info)).ServeHTTP(w, r)
	})
	return mux
}
