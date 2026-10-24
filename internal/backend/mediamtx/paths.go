package mediamtx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

const requestTimeout = 10 * time.Second

type Info struct {
	Started time.Time `json:"started"`
	Version string    `json:"version"`
}

type PathConfig struct {
	Name                       string `json:"name,omitempty"`
	Source                     string `json:"source,omitempty"`
	SourceOnDemand             bool   `json:"sourceOnDemand,omitempty"`
	SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout,omitempty"`
	SourceOnDemandCloseAfter   string `json:"sourceOnDemandCloseAfter,omitempty"`
	RTSPTransport              string `json:"rtspTransport,omitempty"`
}

type Path struct {
	Name          string       `json:"name"`
	ConfigName    string       `json:"confName"`
	Online        bool         `json:"online"`
	OnlineTime    *time.Time   `json:"onlineTime"`
	Available     bool         `json:"available"`
	AvailableTime *time.Time   `json:"availableTime"`
	Source        *PathSource  `json:"source"`
	Readers       []PathReader `json:"readers"`
	Tracks        []PathTrack  `json:"tracks2"`
}

type PathSource struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PathReader struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PathTrack struct {
	Codec      string          `json:"codec"`
	CodecProps json.RawMessage `json:"codecProps"`
}

func (c *Client) GetInfo(ctx context.Context) (Info, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var info Info
	if err := c.do(ctx, http.MethodGet, "/v3/info", nil, nil, &info); err != nil {
		return Info{}, err
	}
	return info, nil
}

func (c *Client) ListPathConfigs(ctx context.Context) ([]PathConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	return listAll[PathConfig](ctx, c, "/v3/config/paths/list")
}

func (c *Client) GetPathConfig(ctx context.Context, name string) (PathConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	endpoint, err := namedPath("/v3/config/paths/get/", name)
	if err != nil {
		return PathConfig{}, err
	}
	var config PathConfig
	if err := c.do(ctx, http.MethodGet, endpoint, nil, nil, &config); err != nil {
		return PathConfig{}, err
	}
	return config, nil
}

func (c *Client) AddPath(ctx context.Context, name string, config PathConfig) error {
	return c.writePath(ctx, "/v3/config/paths/add/", name, config)
}

func (c *Client) ReplacePath(ctx context.Context, name string, config PathConfig) error {
	return c.writePath(ctx, "/v3/config/paths/replace/", name, config)
}

func (c *Client) writePath(ctx context.Context, prefix, name string, config PathConfig) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	endpoint, err := namedPath(prefix, name)
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPost, endpoint, nil, config, nil)
}

func (c *Client) DeletePath(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	endpoint, err := namedPath("/v3/config/paths/delete/", name)
	if err != nil {
		return err
	}
	if err := c.do(ctx, http.MethodDelete, endpoint, nil, nil, nil); !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
}

func (c *Client) ListPaths(ctx context.Context) ([]Path, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	return listAll[Path](ctx, c, "/v3/paths/list")
}

func (c *Client) GetPath(ctx context.Context, name string) (Path, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	endpoint, err := namedPath("/v3/paths/get/", name)
	if err != nil {
		return Path{}, err
	}
	var path Path
	if err := c.do(ctx, http.MethodGet, endpoint, nil, nil, &path); err != nil {
		return Path{}, err
	}
	return path, nil
}

func namedPath(prefix, name string) (string, error) {
	if name == "" {
		return "", errors.New("MediaMTX path name is required")
	}
	return prefix + url.PathEscape(name), nil
}
