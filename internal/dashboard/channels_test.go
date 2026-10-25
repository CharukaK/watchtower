package dashboard

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/CharukaK/watchtower/internal/backend/isapi"
)

func TestChannelFormatting(t *testing.T) {
	if got := formatResolutions("960,640", "1080,720"); got != "960 × 1080, 640 × 720" {
		t.Fatalf("resolutions = %q", got)
	}
	if got := formatFrameRateOptions("1500,1250"); got != "15 fps, 12.5 fps" {
		t.Fatalf("frame rates = %q", got)
	}
	if got := formatOptions("true, false"); got != "true, false" {
		t.Fatalf("options = %q", got)
	}
}

func TestSubstreams(t *testing.T) {
	channels := []isapi.StreamingChannel{
		{ID: 101, Enabled: true},
		{ID: 102, Enabled: true, Video: isapi.StreamingChannelVideo{InputChannelID: 1}},
		{ID: 202, Enabled: false, Video: isapi.StreamingChannelVideo{InputChannelID: 2}},
		{ID: 302, Enabled: true, Video: isapi.StreamingChannelVideo{InputChannelID: 3}},
	}

	streams := substreams(channels)
	if len(streams) != 2 || streams[0].ID != 102 || streams[1].ID != 302 {
		t.Fatalf("substreams = %+v, want enabled channels 102 and 302", streams)
	}
	if got := substreamURL("http://media:8889", 102); got != "http://media:8889/channel-102?controls=true&muted=true&autoplay=true&playsinline=true" {
		t.Fatalf("substream URL = %q", got)
	}

	var rendered bytes.Buffer
	if err := substreamGrid(channels, "http://media:8889").Render(context.Background(), &rendered); err != nil {
		t.Fatal(err)
	}
	html := rendered.String()
	if !strings.Contains(html, "channel-102") || !strings.Contains(html, "channel-302") || strings.Contains(html, "channel-202") {
		t.Fatalf("substream grid = %q", html)
	}
}

func TestCameraViews(t *testing.T) {
	channels := []isapi.StreamingChannel{
		{ID: 101, Name: "Gate", Enabled: true, Video: isapi.StreamingChannelVideo{InputChannelID: 1, CodecType: "H.265", ResolutionWidth: 1920, ResolutionHeight: 1080}},
		{ID: 102, Name: "Gate", Enabled: true, Video: isapi.StreamingChannelVideo{InputChannelID: 1, CodecType: "H.264", ResolutionWidth: 640, ResolutionHeight: 360}},
		{ID: 103, Name: "Gate", Enabled: true, Video: isapi.StreamingChannelVideo{InputChannelID: 1, CodecType: "H.264", ResolutionWidth: 1280, ResolutionHeight: 720}},
		{ID: 201, Name: "201", Enabled: true, Video: isapi.StreamingChannelVideo{InputChannelID: 2, CodecType: "H.264", ResolutionWidth: 1280, ResolutionHeight: 720}},
		{ID: 302, Name: "Warehouse", Enabled: false, Video: isapi.StreamingChannelVideo{InputChannelID: 3}},
	}

	if got := playerURL(102, false); got != "/player/102?controls=false&muted=true&autoplay=true&playsinline=true" {
		t.Fatalf("player URL = %q", got)
	}

	cameras := cameraViews(channels)
	if len(cameras) != 2 {
		t.Fatalf("got %d cameras, want 2", len(cameras))
	}
	if cameras[0].Name != "Gate" || cameras[0].PreviewStreamID != 102 || cameras[0].PreviewStreamURL != playerURL(102, false) || cameras[0].PreviewCodec != "H.264" {
		t.Fatalf("first camera = %+v, want Gate preview using H.264 substream", cameras[0])
	}
	if cameras[0].StreamID != 101 || cameras[0].StreamURL != playerURL(101, true) || cameras[0].Codec != "H.265" {
		t.Fatalf("first camera = %+v, want detail using main stream 101", cameras[0])
	}
	if cameras[1].Name != "Camera 2" || cameras[1].StreamID != 201 {
		t.Fatalf("second camera = %+v, want fallback camera name and enabled stream", cameras[1])
	}
}

func TestGroupChannels(t *testing.T) {
	channels := []isapi.StreamingChannel{
		{Name: "201", Video: isapi.StreamingChannelVideo{InputChannelID: 2}},
		{Name: "101", Video: isapi.StreamingChannelVideo{InputChannelID: 1}},
		{Name: "202", Video: isapi.StreamingChannelVideo{InputChannelID: 2}},
	}

	groups := groupChannels(channels)
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(groups))
	}
	if groups[0].InputChannelID != 2 || len(groups[0].Channels) != 2 {
		t.Fatalf("first group = %+v, want input channel 2 with two streams", groups[0])
	}
	if groups[1].InputChannelID != 1 || len(groups[1].Channels) != 1 {
		t.Fatalf("second group = %+v, want input channel 1 with one stream", groups[1])
	}
}
