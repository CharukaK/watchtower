package isapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/icholy/digest"
)

var ErrUnauthorized = errors.New("hikvision: unauthorized")

type Client struct {
	baseURL  *url.URL
	username string
	password string
	http     *http.Client
}

func (c *Client) endpoint(relativePath string) string {
	u := *c.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/" +
		strings.TrimLeft(relativePath, "/")
	return u.String()
}

func (c *Client) StreamURL(channelID int) string {
	u := *c.baseURL
	u.Scheme = "rtsp"
	u.Path = fmt.Sprintf("/Streaming/Channels/%d", channelID)
	u.User = url.UserPassword(c.username, c.password)
	return u.String()
}

func (c *Client) do(
	ctx context.Context,
	method string,
	relativePath string,
	body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		c.endpoint(relativePath),
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send ISAPI request: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, ErrUnauthorized
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("ISAPI request returned %s", resp.Status)
	}

	return resp, nil
}

func NewClient(deviceURL, username, password string) (*Client, error) {
	c := &Client{}

	if len(deviceURL) == 0 || len(username) == 0 || len(password) == 0 {
		return nil, fmt.Errorf("missing information, please provide host, username and password")
	}

	if du, err := url.Parse(deviceURL); err != nil {
		return nil, err
	} else {
		c.baseURL = du
	}

	if c.baseURL.User != nil {
		return nil, errors.New("put credentials in NewClient, not the URL")
	}

	if c.baseURL.RawQuery != "" || c.baseURL.Fragment != "" {
		return nil, errors.New("device URL must not contain a query or fragment")
	}

	if c.baseURL.Path != "" && c.baseURL.Path != "/" {
		return nil, errors.New("device URL must not contain a path")
	}

	c.baseURL.Path = "/ISAPI"
	c.username = username
	c.password = password

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.ResponseHeaderTimeout = 10 * time.Second

	authTransport := &digest.Transport{
		Username:  username,
		Password:  password,
		Transport: transport,
		Probe:     true,
	}

	c.http = &http.Client{
		Transport: authTransport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return c, nil
}
