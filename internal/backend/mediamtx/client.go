package mediamtx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("mediamtx: not found")
	ErrUnauthorized = errors.New("mediamtx: unauthorized")
)

type Client struct {
	baseURL       string
	webRTCBaseURL string
	username      string
	password      string
	http          *http.Client
}

func NewClient(apiURL, username, password string) (*Client, error) {
	if apiURL == "" || username == "" || password == "" {
		return nil, errors.New("MediaMTX API URL, username, and password are required")
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse MediaMTX API URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, errors.New("MediaMTX API URL must be an HTTP or HTTPS URL")
	}
	if u.User != nil {
		return nil, errors.New("put credentials in NewClient, not the URL")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("MediaMTX API URL must not contain a query or fragment")
	}
	if u.Path != "" && u.Path != "/" {
		return nil, errors.New("MediaMTX API URL must not contain a path")
	}

	webRTCURL := *u
	webRTCURL.Host = net.JoinHostPort(u.Hostname(), "8889")
	webRTCURL.Path = ""

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.ResponseHeaderTimeout = 10 * time.Second

	return &Client{
		baseURL:       strings.TrimRight(u.String(), "/"),
		webRTCBaseURL: strings.TrimRight(webRTCURL.String(), "/"),
		username:      username,
		password:      password,
		http: &http.Client{
			Transport: transport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (c *Client) WebRTCBaseURL() string {
	return c.webRTCBaseURL
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, input, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode MediaMTX request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	endpoint := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(query) != 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("create MediaMTX request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send MediaMTX request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("MediaMTX API request returned %s", resp.Status)
	}

	reader := io.LimitReader(resp.Body, 1<<20)
	if output == nil {
		_, _ = io.Copy(io.Discard, reader)
		return nil
	}
	if err := json.NewDecoder(reader).Decode(output); err != nil {
		return fmt.Errorf("decode MediaMTX response: %w", err)
	}
	return nil
}

type listResponse[T any] struct {
	ItemCount int `json:"itemCount"`
	PageCount int `json:"pageCount"`
	Items     []T `json:"items"`
}

func listAll[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	var items []T
	for pageNumber := 0; ; pageNumber++ {
		var page listResponse[T]
		query := url.Values{"page": {fmt.Sprint(pageNumber)}}
		if err := c.do(ctx, http.MethodGet, path, query, nil, &page); err != nil {
			return nil, err
		}
		items = append(items, page.Items...)
		if pageNumber+1 >= page.PageCount {
			return items, nil
		}
	}
}
