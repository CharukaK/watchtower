package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/CharukaK/watchtower/internal/app"
	"github.com/CharukaK/watchtower/internal/backend/isapi"
	"github.com/CharukaK/watchtower/internal/backend/mediamtx"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	hostUrl := os.Getenv("HOST_URL")
	userName := os.Getenv("USER")
	password := os.Getenv("PASSWORD")

	mediamtxHost := os.Getenv("MEDIAMTX_HOST")
	mediamtxUser := os.Getenv("MEDIAMTX_USER")
	mediamtxPass := os.Getenv("MEDIAMTX_PASSWORD")

	isAPIClient, err := isapi.NewClient(hostUrl, userName, password)
	if err != nil {
		slog.Error("invalid configuration provided", "error", err)
		os.Exit(1)
	}

	mediamtxClient, err := mediamtx.NewClient(mediamtxHost, mediamtxUser, mediamtxPass)
	if err != nil {
		slog.Error("invalid configuration provided", "error", err)
		os.Exit(1)
	}

	if err := app.RegisterStreams(context.Background(), isAPIClient, mediamtxClient); err != nil {
		slog.Error("failed to register streams", "error", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app.New(isAPIClient, mediamtxClient),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
