package isapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetAdminAccesses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Security/adminAccesses" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/Security/adminAccesses")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8" ?>
<AdminAccessProtocolList version="1.0" xmlns="http://www.hikvision.com/ver20/XMLSchema">
<AdminAccessProtocol version="1.0"><id>1</id><enabled>true</enabled><protocol>HTTP</protocol><portNo>80</portNo></AdminAccessProtocol>
<AdminAccessProtocol version="1.0"><id>3</id><enabled>false</enabled><protocol>HTTPS</protocol><redirectToHttps>false</redirectToHttps><portNo>443</portNo></AdminAccessProtocol>
<AdminAccessProtocol version="1.0"><id>4</id><protocol>DEV_MANAGE</protocol><portNo>8000</portNo></AdminAccessProtocol>
</AdminAccessProtocolList>`))
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := &Client{baseURL: baseURL, http: server.Client()}

	protocols, err := client.GetAdminAccesses(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(protocols) != 3 {
		t.Fatalf("protocol count = %d, want 3", len(protocols))
	}
	if protocols[0].Version != "1.0" || protocols[0].ID != 1 || protocols[0].Enabled == nil || !*protocols[0].Enabled || protocols[0].Protocol != "HTTP" || protocols[0].Port != 80 {
		t.Fatalf("HTTP protocol = %+v", protocols[0])
	}
	if protocols[1].Enabled == nil || *protocols[1].Enabled || protocols[1].RedirectToHTTPS == nil || *protocols[1].RedirectToHTTPS {
		t.Fatalf("HTTPS protocol = %+v", protocols[1])
	}
	if protocols[2].Enabled != nil || protocols[2].RedirectToHTTPS != nil || protocols[2].Port != 8000 {
		t.Fatalf("DEV_MANAGE protocol = %+v", protocols[2])
	}
}
