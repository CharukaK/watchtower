package dashboard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CharukaK/watchtower/internal/backend/isapi"
)

type channelGroup struct {
	InputChannelID int
	Channels       []isapi.StreamingChannel
}

type cameraView struct {
	InputChannelID   int
	Name             string
	PreviewStreamID  int
	PreviewStreamURL string
	PreviewCodec     string
	StreamID         int
	StreamURL        string
	SnapshotURL      string
	Codec            string
	Resolution       string
}

func formatOptions(options string) string {
	values := strings.Split(options, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	return strings.Join(values, ", ")
}

func formatResolutions(widths, heights string) string {
	widthOptions := strings.Split(widths, ",")
	heightOptions := strings.Split(heights, ",")
	if len(widthOptions) != len(heightOptions) {
		return formatOptions(widths) + " × " + formatOptions(heights)
	}

	resolutions := make([]string, len(widthOptions))
	for i := range widthOptions {
		resolutions[i] = strings.TrimSpace(widthOptions[i]) + " × " + strings.TrimSpace(heightOptions[i])
	}
	return strings.Join(resolutions, ", ")
}

func formatFrameRate(frameRate int) string {
	return fmt.Sprintf("%g fps", float64(frameRate)/100)
}

func formatFrameRateOptions(options string) string {
	values := strings.Split(options, ",")
	for i, value := range values {
		frameRate, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return formatOptions(options)
		}
		values[i] = formatFrameRate(frameRate)
	}
	return strings.Join(values, ", ")
}

func formatTransports(protocols []isapi.StreamingControlProtocol) string {
	values := make([]string, len(protocols))
	for i, protocol := range protocols {
		values[i] = protocol.Transport
	}
	return strings.Join(values, ", ")
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

func substreams(channels []isapi.StreamingChannel) []isapi.StreamingChannel {
	streams := make([]isapi.StreamingChannel, 0)
	for _, channel := range channels {
		if channel.Enabled && channel.ID%100 == 2 {
			streams = append(streams, channel)
		}
	}
	return streams
}

func substreamURL(baseURL string, channelID int) string {
	return fmt.Sprintf("%s/channel-%d?controls=true&muted=true&autoplay=true&playsinline=true", baseURL, channelID)
}

func playerURL(channelID int, controls bool) string {
	return fmt.Sprintf("/player/%d?controls=%t&muted=true&autoplay=true&playsinline=true", channelID, controls)
}

func cameraViews(channels []isapi.StreamingChannel) []cameraView {
	cameras := make([]cameraView, 0)
	for _, group := range groupChannels(channels) {
		var preview, detail *isapi.StreamingChannel
		for i := range group.Channels {
			channel := &group.Channels[i]
			if !channel.Enabled {
				continue
			}

			channelPixels := channel.Video.ResolutionWidth * channel.Video.ResolutionHeight
			if detail == nil || channelPixels > detail.Video.ResolutionWidth*detail.Video.ResolutionHeight {
				detail = channel
			}
			if preview == nil {
				preview = channel
				continue
			}

			previewH264 := strings.EqualFold(preview.Video.CodecType, "H.264")
			channelH264 := strings.EqualFold(channel.Video.CodecType, "H.264")
			previewPixels := preview.Video.ResolutionWidth * preview.Video.ResolutionHeight
			if channelH264 && (!previewH264 || channelPixels > 0 && (previewPixels == 0 || channelPixels < previewPixels)) {
				preview = channel
			}
		}
		if preview == nil || detail == nil {
			continue
		}

		name := fmt.Sprintf("Camera %d", group.InputChannelID)
		for _, channel := range group.Channels {
			candidate := strings.TrimSpace(channel.Name)
			if candidate != "" && candidate != strconv.Itoa(channel.ID) {
				name = candidate
				break
			}
		}

		cameras = append(cameras, cameraView{
			InputChannelID:   group.InputChannelID,
			Name:             name,
			PreviewStreamID:  preview.ID,
			PreviewStreamURL: playerURL(preview.ID, false),
			PreviewCodec:     preview.Video.CodecType,
			StreamID:         detail.ID,
			StreamURL:        playerURL(detail.ID, true),
			SnapshotURL:      fmt.Sprintf("/channels/%d/picture", detail.ID),
			Codec:            detail.Video.CodecType,
			Resolution:       fmt.Sprintf("%d × %d", detail.Video.ResolutionWidth, detail.Video.ResolutionHeight),
		})
	}
	return cameras
}

func groupChannels(channels []isapi.StreamingChannel) []channelGroup {
	groups := make([]channelGroup, 0)
	indexes := make(map[int]int)

	for _, channel := range channels {
		inputChannelID := channel.Video.InputChannelID
		index, exists := indexes[inputChannelID]
		if !exists {
			index = len(groups)
			indexes[inputChannelID] = index
			groups = append(groups, channelGroup{InputChannelID: inputChannelID})
		}
		groups[index].Channels = append(groups[index].Channels, channel)
	}

	return groups
}
