package app

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/CharukaK/watchtower/internal/backend/isapi"
	"github.com/CharukaK/watchtower/internal/backend/mediamtx"
	"github.com/CharukaK/watchtower/internal/dashboard"
	"github.com/CharukaK/watchtower/web"
	"github.com/a-h/templ"
)

func New(isapiClient *isapi.Client, mediamtxClient *mediamtx.Client) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServer(http.FS(web.Static)))
	mux.HandleFunc("GET /channels/{id}/capabilities", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if channelID, err := strconv.Atoi(id); err != nil || channelID < 1 {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}

		capability, err := isapiClient.GetChannelCapabilities(r.Context(), id)
		if err != nil {
			slog.WarnContext(r.Context(), "failed to get channel capabilities", "channel", id, "error", err)
			http.Error(w, "capabilities unavailable", http.StatusBadGateway)
			return
		}

		w.Header().Set("Datastar-Selector", "#channel-"+id+"-capabilities")
		w.Header().Set("Datastar-Mode", "inner")
		templ.Handler(dashboard.ChannelCapabilities(capability)).ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /channels/{id}/picture", isapiClient.GetPicture)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		mediamtxInfo, err := mediamtxClient.GetInfo(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get MediaMTX information", "error", err)
			http.Error(w, "MediaMTX unavailable", http.StatusBadGateway)
			return
		}

		info, err := isapiClient.GetDeviceInfo(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get device information", "error", err)
			http.Error(w, "device unavailable", http.StatusBadGateway)
			return
		}

		channels, err := isapiClient.GetChannels(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get channels information", "error", err)
			http.Error(w, "error fetching channels", http.StatusBadGateway)
			return
		}

		templ.Handler(dashboard.Dashboard(info, mediamtxInfo, channels, mediamtxClient.WebRTCBaseURL())).ServeHTTP(w, r)
	})
	return mux
}
