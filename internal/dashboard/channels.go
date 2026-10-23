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
