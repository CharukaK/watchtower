package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	tests := []struct {
		path         string
		status       int
		bodyContains string
	}{
		{path: "/", status: http.StatusOK, bodyContains: "/static/datastar.js"},
		{path: "/static/datastar.js", status: http.StatusOK, bodyContains: "Datastar"},
		{path: "/missing", status: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()
			New().ServeHTTP(response, request)

			result := response.Result()
			defer result.Body.Close()
			if result.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d", result.StatusCode, tt.status)
			}

			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatal(err)
			}
			if tt.bodyContains != "" && !strings.Contains(string(body), tt.bodyContains) {
				t.Fatalf("body does not contain %q", tt.bodyContains)
			}
		})
	}
}
