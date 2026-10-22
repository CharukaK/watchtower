package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/CharukaK/watchtower/internal/app"
	"github.com/CharukaK/watchtower/internal/backend/isapi"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	hostUrl := os.Getenv("HOST_URL")
	userName := os.Getenv("USER")
	password := os.Getenv("PASSWORD")

	isAPIClient, err := isapi.NewClient(hostUrl, userName, password)
	if err != nil {
		slog.Error("invalid configuration provided", "error", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app.New(isAPIClient),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
