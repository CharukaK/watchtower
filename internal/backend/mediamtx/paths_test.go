package mediamtx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientAPI(t *testing.T) {
	var configListRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "apiuser" || password != "secret" {
			t.Errorf("basic auth = %q, %q, %t", username, password, ok)
		}
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /v3/info":
			_, _ = w.Write([]byte(`{"started":"2026-10-20T10:00:00Z","version":"v1.21.2"}`))
		case "GET /v3/config/paths/list":
			configListRequests++
			switch r.URL.Query().Get("page") {
			case "0":
				_, _ = w.Write([]byte(`{"itemCount":2,"pageCount":2,"items":[{"name":"camera-1","source":"rtsp://user:pass@dvr/Streaming/Channels/102","sourceOnDemand":true,"rtspTransport":"tcp"}]}`))
			case "1":
				_, _ = w.Write([]byte(`{"itemCount":2,"pageCount":2,"items":[{"name":"camera-2","source":"rtsp://user:pass@dvr/Streaming/Channels/202","sourceOnDemand":true,"rtspTransport":"tcp"}]}`))
			default:
				t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			}
		case "GET /v3/config/paths/get/camera one":
			if r.URL.EscapedPath() != "/v3/config/paths/get/camera%20one" {
				t.Errorf("escaped path = %q", r.URL.EscapedPath())
			}
			_, _ = w.Write([]byte(`{"name":"camera one","source":"rtsp://dvr/stream","sourceOnDemand":true}`))
		case "POST /v3/config/paths/add/camera-1":
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("content type = %q", r.Header.Get("Content-Type"))
			}
			var config PathConfig
			if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
				t.Error(err)
			}
			if config.Source != "rtsp://user:pass@dvr/Streaming/Channels/102" || !config.SourceOnDemand || config.SourceOnDemandStartTimeout != "10s" || config.SourceOnDemandCloseAfter != "10s" || config.RTSPTransport != "tcp" {
				t.Errorf("path config = %+v", config)
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "DELETE /v3/config/paths/delete/missing":
			http.Error(w, `{"status":"error","error":"not found"}`, http.StatusNotFound)
		case "GET /v3/paths/list":
			_, _ = w.Write([]byte(`{"itemCount":1,"pageCount":1,"items":[{"name":"camera-1","online":true,"available":true,"tracks2":[{"codec":"H264"}]}]}`))
		case "GET /v3/paths/get/camera-1":
			_, _ = w.Write([]byte(`{"name":"camera-1","confName":"camera-1","online":true,"onlineTime":"2026-10-20T10:01:00Z","available":true,"availableTime":"2026-10-20T10:01:01Z","source":{"id":"source-id","type":"rtspSource"},"readers":[{"id":"reader-id","type":"webRTCSession"}],"tracks2":[{"codec":"H264","codecProps":{"width":1920,"height":1080}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "apiuser", "secret")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	info, err := client.GetInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "v1.21.2" || !info.Started.Equal(time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("info = %+v", info)
	}

	configs, err := client.ListPathConfigs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 2 || configs[0].Name != "camera-1" || configs[1].Name != "camera-2" || configListRequests != 2 {
		t.Fatalf("configs = %+v, requests = %d", configs, configListRequests)
	}

	config, err := client.GetPathConfig(ctx, "camera one")
	if err != nil {
		t.Fatal(err)
	}
	if config.Name != "camera one" || !config.SourceOnDemand {
		t.Fatalf("config = %+v", config)
	}

	desired := PathConfig{
		Source:                     "rtsp://user:pass@dvr/Streaming/Channels/102",
		SourceOnDemand:             true,
		SourceOnDemandStartTimeout: "10s",
		SourceOnDemandCloseAfter:   "10s",
		RTSPTransport:              "tcp",
	}
	if err := client.AddPath(ctx, "camera-1", desired); err != nil {
		t.Fatal(err)
	}
	if err := client.DeletePath(ctx, "missing"); err != nil {
		t.Fatalf("delete missing path: %v", err)
	}

	paths, err := client.ListPaths(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0].Name != "camera-1" || !paths[0].Online || paths[0].Tracks[0].Codec != "H264" {
		t.Fatalf("paths = %+v", paths)
	}

	path, err := client.GetPath(ctx, "camera-1")
	if err != nil {
		t.Fatal(err)
	}
	if path.Source == nil || path.Source.Type != "rtspSource" || len(path.Readers) != 1 || path.Readers[0].Type != "webRTCSession" || !strings.Contains(string(path.Tracks[0].CodecProps), `"width":1920`) {
		t.Fatalf("path = %+v", path)
	}
}

func TestClientErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/info":
			w.WriteHeader(http.StatusUnauthorized)
		case "/v3/config/paths/get/missing":
			w.WriteHeader(http.StatusNotFound)
		case "/v3/config/paths/add/camera-1":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "apiuser", "secret")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.GetInfo(context.Background()); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("GetInfo error = %v", err)
	}
	if _, err := client.GetPathConfig(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetPathConfig error = %v", err)
	}

	secretSource := "rtsp://admin:camera-password@dvr/stream"
	err = client.AddPath(context.Background(), "camera-1", PathConfig{Source: secretSource})
	if err == nil || strings.Contains(err.Error(), secretSource) || strings.Contains(err.Error(), "camera-password") {
		t.Fatalf("AddPath error = %v", err)
	}
}

func TestNewClientRejectsInvalidURL(t *testing.T) {
	for _, apiURL := range []string{"", "localhost:9997", "ftp://localhost", "http://user:pass@localhost:9997", "http://localhost:9997/v3"} {
		t.Run(apiURL, func(t *testing.T) {
			if _, err := NewClient(apiURL, "apiuser", "secret"); err == nil {
				t.Fatalf("NewClient(%q) succeeded", apiURL)
			}
		})
	}
}
