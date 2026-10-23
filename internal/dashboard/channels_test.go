package dashboard

import (
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
