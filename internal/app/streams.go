package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CharukaK/watchtower/internal/backend/isapi"
	"github.com/CharukaK/watchtower/internal/backend/mediamtx"
)

func RegisterStreams(ctx context.Context, isapiClient *isapi.Client, mediamtxClient *mediamtx.Client) error {
	channels, err := isapiClient.GetChannels(ctx)
	if err != nil {
		return fmt.Errorf("get channels: %w", err)
	}

	configs, err := mediamtxClient.ListPathConfigs(ctx)
	if err != nil {
		return fmt.Errorf("list MediaMTX paths: %w", err)
	}
	existing := make(map[string]bool, len(configs))
	for _, config := range configs {
		existing[config.Name] = true
	}

	for _, channel := range channels {
		name := "channel-" + strconv.Itoa(channel.ID)
		if !channel.Enabled || existing[name] {
			continue
		}

		err := mediamtxClient.AddPath(ctx, name, mediamtx.PathConfig{
			Source:                     isapiClient.StreamURL(channel.ID),
			SourceOnDemand:             true,
			SourceOnDemandStartTimeout: "10s",
			SourceOnDemandCloseAfter:   "10s",
			RTSPTransport:              "tcp",
		})
		if err != nil {
			return fmt.Errorf("register stream %s: %w", name, err)
		}
	}

	return nil
}
