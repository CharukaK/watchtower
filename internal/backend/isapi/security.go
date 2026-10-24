package isapi

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AdminAccessProtocolList struct {
	XMLName   xml.Name              `xml:"AdminAccessProtocolList"`
	Version   string                `xml:"version,attr"`
	Protocols []AdminAccessProtocol `xml:"AdminAccessProtocol"`
}

type AdminAccessProtocol struct {
	Version         string `xml:"version,attr"`
	ID              int    `xml:"id"`
	Enabled         *bool  `xml:"enabled"`
	Protocol        string `xml:"protocol"`
	RedirectToHTTPS *bool  `xml:"redirectToHttps"`
	Port            int    `xml:"portNo"`
}

func (c *Client) GetAdminAccesses(ctx context.Context) ([]AdminAccessProtocol, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.do(ctx, http.MethodGet, "/Security/adminAccesses", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result AdminAccessProtocolList
	reader := io.LimitReader(resp.Body, 1<<20)
	if err := xml.NewDecoder(reader).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode admin accesses: %w", err)
	}

	return result.Protocols, nil
}
