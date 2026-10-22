package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/CharukaK/watchtower/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
