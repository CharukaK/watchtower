package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CharukaK/watchtower/internal/backend/isapi"
	"github.com/CharukaK/watchtower/internal/backend/mediamtx"
)

func TestRegisterStreams(t *testing.T) {
	camera := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ISAPI/Streaming/channels" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `<StreamingChannelList>
			<StreamingChannel><id>101</id><enabled>true</enabled></StreamingChannel>
			<StreamingChannel><id>102</id><enabled>true</enabled></StreamingChannel>
			<StreamingChannel><id>103</id><enabled>false</enabled></StreamingChannel>
		</StreamingChannelList>`)
	}))
	defer camera.Close()

	var addedName string
	var addedConfig mediamtx.PathConfig
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "media-user" || password != "media-pass" {
			t.Errorf("basic auth = %q, %q, %t", username, password, ok)
		}

		switch r.Method + " " + r.URL.Path {
		case "GET /v3/config/paths/list":
			_, _ = io.WriteString(w, `{"itemCount":1,"pageCount":1,"items":[{"name":"channel-102"}]}`)
		case "POST /v3/config/paths/add/channel-101":
			addedName = strings.TrimPrefix(r.URL.Path, "/v3/config/paths/add/")
			if err := json.NewDecoder(r.Body).Decode(&addedConfig); err != nil {
				t.Error(err)
			}
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer media.Close()

	cameraClient, err := isapi.NewClient(camera.URL, "camera-user", "camera-pass")
	if err != nil {
		t.Fatal(err)
	}
	mediaClient, err := mediamtx.NewClient(media.URL, "media-user", "media-pass")
	if err != nil {
		t.Fatal(err)
	}

	if err := RegisterStreams(context.Background(), cameraClient, mediaClient); err != nil {
		t.Fatal(err)
	}

	wantSource := strings.Replace(camera.URL, "http://", "rtsp://camera-user:camera-pass@", 1) + "/Streaming/Channels/101"
	if addedName != "channel-101" {
		t.Fatalf("added path = %q, want channel-101", addedName)
	}
	if addedConfig.Source != wantSource || !addedConfig.SourceOnDemand || addedConfig.SourceOnDemandStartTimeout != "10s" || addedConfig.SourceOnDemandCloseAfter != "10s" || addedConfig.RTSPTransport != "tcp" {
		t.Fatalf("added config = %+v, want source %q with on-demand TCP settings", addedConfig, wantSource)
	}
}
