package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/CharukaK/watchtower/internal/backend/isapi"
	"github.com/CharukaK/watchtower/internal/backend/mediamtx"
)

func testISAPIClient(t *testing.T, capabilitiesStatus int) (*isapi.Client, *atomic.Int32) {
	t.Helper()

	var capabilitiesRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		switch r.URL.Path {
		case "/ISAPI/System/deviceInfo":
			_, _ = io.WriteString(w, `<DeviceInfo><deviceName>Recorder</deviceName><model>DS-7600</model><serialNumber>123</serialNumber><firmwareVersion>1.0</firmwareVersion></DeviceInfo>`)
		case "/ISAPI/Streaming/channels":
			_, _ = io.WriteString(w, `<StreamingChannelList><StreamingChannel><id>101</id><channelName>101</channelName><enabled>true</enabled><Transport><ControlProtocolList><ControlProtocol><streamingTransport>RTSP</streamingTransport></ControlProtocol></ControlProtocolList></Transport><Video><videoInputChannelID>1</videoInputChannelID><videoCodecType>H.265</videoCodecType><videoResolutionWidth>960</videoResolutionWidth><videoResolutionHeight>1080</videoResolutionHeight><videoQualityControlType>VBR</videoQualityControlType><vbrUpperCap>1024</vbrUpperCap><vbrLowerCap>32</vbrLowerCap><maxFrameRate>1500</maxFrameRate><snapShotImageType>JPEG</snapShotImageType><GovLength>15</GovLength></Video><Audio><enabled>true</enabled><audioInputChannelID>1</audioInputChannelID><audioCompressionType>G.711ulaw</audioCompressionType></Audio></StreamingChannel></StreamingChannelList>`)
		case "/ISAPI/Streaming/channels/101/capabilities":
			capabilitiesRequests.Add(1)
			if capabilitiesStatus != http.StatusOK {
				w.WriteHeader(capabilitiesStatus)
				return
			}
			_, _ = io.WriteString(w, `<StreamingChannel><id>101</id><Video><videoCodecType opt="H.264,H.265">H.265</videoCodecType><videoResolutionWidth opt="960,640">960</videoResolutionWidth><videoResolutionHeight opt="1080,720">1080</videoResolutionHeight><videoQualityControlType opt="CBR,VBR">VBR</videoQualityControlType><fixedQuality opt="20,60">60</fixedQuality><constantBitRate min="32" max="4096">0</constantBitRate><maxFrameRate opt="1500,1200">1500</maxFrameRate><SVC><SVCMode opt="manual,auto">manual</SVCMode></SVC></Video><isSupportPictureByUrl>true</isSupportPictureByUrl></StreamingChannel>`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := isapi.NewClient(server.URL, "user", "password")
	if err != nil {
		t.Fatal(err)
	}
	return client, &capabilitiesRequests
}

func testMediaMTXClient(t *testing.T, status int) (*mediamtx.Client, *atomic.Int32) {
	t.Helper()

	var infoRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		infoRequests.Add(1)
		if r.URL.Path != "/v3/info" {
			http.NotFound(w, r)
			return
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_, _ = io.WriteString(w, `{"started":"2026-10-20T10:00:00Z","version":"v1.21.2"}`)
	}))
	t.Cleanup(server.Close)

	client, err := mediamtx.NewClient(server.URL, "user", "password")
	if err != nil {
		t.Fatal(err)
	}
	return client, &infoRequests
}

func TestServer(t *testing.T) {
	client, capabilitiesRequests := testISAPIClient(t, http.StatusOK)
	mediamtxClient, infoRequests := testMediaMTXClient(t, http.StatusOK)
	server := New(client, mediamtxClient)
	tests := []struct {
		path         string
		status       int
		bodyContains []string
	}{
		{path: "/", status: http.StatusOK, bodyContains: []string{"Live cameras", "Recent events", "/player/101?controls=true", "data-signals:screen=", "data-on:toggle=", "MediaMTX", "v1.21.2", "Tue, 20 Oct 2026 10:00:00 UTC"}},
		{path: "/player/101", status: http.StatusOK, bodyContains: []string{"object-fit: fill", `data-channel-id="101"`, "/static/player.js"}},
		{path: "/player/nope", status: http.StatusBadRequest, bodyContains: []string{"invalid channel id"}},
		{path: "/static/player.js", status: http.StatusOK, bodyContains: []string{"MediaMTXWebRTCReader", "/media/channel-"}},
		{path: "/static/datastar.js", status: http.StatusOK, bodyContains: []string{"Datastar"}},
		{path: "/missing", status: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			result := response.Result()
			defer result.Body.Close()
			if result.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d", result.StatusCode, tt.status)
			}

			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tt.bodyContains {
				if !strings.Contains(string(body), want) {
					t.Fatalf("body does not contain %q", want)
				}
			}
		})
	}

	if got := capabilitiesRequests.Load(); got != 0 {
		t.Fatalf("initial dashboard made %d capabilities requests", got)
	}
	if got := infoRequests.Load(); got != 1 {
		t.Fatalf("initial dashboard made %d MediaMTX info requests, want 1", got)
	}

	request := httptest.NewRequest(http.MethodGet, "/channels/101/capabilities", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("capabilities status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "H.264, H.265") {
		t.Fatal("capabilities response does not contain codec options")
	}
	if got := response.Header().Get("Datastar-Selector"); got != "#channel-101-capabilities" {
		t.Fatalf("Datastar selector = %q", got)
	}
	if got := response.Header().Get("Datastar-Mode"); got != "inner" {
		t.Fatalf("Datastar mode = %q", got)
	}
	if got := capabilitiesRequests.Load(); got != 1 {
		t.Fatalf("expanded channel made %d capabilities requests, want 1", got)
	}
}

func TestMediaProxy(t *testing.T) {
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "http://mediamtx:8889/channel-101/whep/session-id")
		_, _ = io.WriteString(w, r.Method+" "+r.URL.RequestURI())
	}))
	t.Cleanup(media.Close)

	target, err := url.Parse(media.URL)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPatch, "/media/channel-101/whep?session=1", nil)
	response := httptest.NewRecorder()
	newHandler(nil, nil, target).ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "PATCH /channel-101/whep?session=1" {
		t.Fatalf("proxy response = %d %q", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Location"); got != "/media/channel-101/whep/session-id" {
		t.Fatalf("proxy Location = %q", got)
	}
}

func TestMediaMTXUnavailable(t *testing.T) {
	client, _ := testISAPIClient(t, http.StatusOK)
	mediamtxClient, _ := testMediaMTXClient(t, http.StatusInternalServerError)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	New(client, mediamtxClient).ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadGateway)
	}
	if !strings.Contains(response.Body.String(), "MediaMTX unavailable") {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestChannelCapabilitiesFailure(t *testing.T) {
	client, _ := testISAPIClient(t, http.StatusInternalServerError)
	mediamtxClient, _ := testMediaMTXClient(t, http.StatusOK)
	request := httptest.NewRequest(http.MethodGet, "/channels/101/capabilities", nil)
	response := httptest.NewRecorder()
	New(client, mediamtxClient).ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadGateway)
	}
}
