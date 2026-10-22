package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CharukaK/watchtower/internal/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}
