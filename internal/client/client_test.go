package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/codeownersnet/atlas/internal/auth"
	"github.com/rs/zerolog"
)

func TestNewClient(t *testing.T) {
	auth, _ := auth.NewBasicAuth("user@example.com", "token123")
	logger := zerolog.Nop()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				BaseURL: "https://example.atlassian.net",
				Auth:    auth,
				Logger:  &logger,
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: &Config{
				Auth:   auth,
				Logger: &logger,
			},
			wantErr: true,
		},
		{
			name: "missing auth",
			config: &Config{
				BaseURL: "https://example.atlassian.net",
				Logger:  &logger,
			},
			wantErr: true,
		},
		{
			name: "with custom headers",
			config: &Config{
				BaseURL: "https://example.atlassian.net",
				Auth:    auth,
				CustomHeaders: map[string]string{
					"X-Custom": "value",
				},
				Logger: &logger,
			},
			wantErr: false,
		},
		{
			name: "with SSL verify disabled",
			config: &Config{
				BaseURL:   "https://example.atlassian.net",
				Auth:      auth,
				SSLVerify: false,
				Logger:    &logger,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestClientGet(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("Expected path /test, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	auth, _ := auth.NewBasicAuth("user@example.com", "token123")
	logger := zerolog.Nop()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    auth,
		Logger:  &logger,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ok") {
		t.Errorf("Expected response to contain 'ok', got %s", string(body))
	}
}

func TestClientPost(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "test") {
			t.Errorf("Expected body to contain 'test', got %s", string(body))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"created": true}`))
	}))
	defer server.Close()

	auth, _ := auth.NewBasicAuth("user@example.com", "token123")
	logger := zerolog.Nop()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    auth,
		Logger:  &logger,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	requestBody := []byte(`{"name": "test"}`)
	resp, err := client.Post(ctx, "/test", requestBody)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0

	// Create a test server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	auth, _ := auth.NewBasicAuth("user@example.com", "token123")
	logger := zerolog.Nop()

	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		Auth:       auth,
		Logger:     &logger,
		MaxRetries: 3,
		RetryDelay: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClientCustomHeaders(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader != "custom-value" {
			t.Errorf("Expected X-Custom-Header: custom-value, got %s", customHeader)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	auth, _ := auth.NewBasicAuth("user@example.com", "token123")
	logger := zerolog.Nop()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    auth,
		CustomHeaders: map[string]string{
			"X-Custom-Header": "custom-value",
		},
		Logger: &logger,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()
}

func TestShouldRetry(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"rate limit", http.StatusTooManyRequests, true},
		{"service unavailable", http.StatusServiceUnavailable, true},
		{"gateway timeout", http.StatusGatewayTimeout, true},
		{"internal server error", http.StatusInternalServerError, true},
		{"bad gateway", http.StatusBadGateway, true},
		{"ok", http.StatusOK, false},
		{"created", http.StatusCreated, false},
		{"bad request", http.StatusBadRequest, false},
		{"unauthorized", http.StatusUnauthorized, false},
		{"forbidden", http.StatusForbidden, false},
		{"not found", http.StatusNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.shouldRetry(tt.statusCode)
			if got != tt.want {
				t.Errorf("shouldRetry(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestShouldBypassProxy(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		noProxy string
		want    bool
	}{
		{
			name:    "exact match",
			host:    "example.com",
			noProxy: "example.com,other.com",
			want:    true,
		},
		{
			name:    "no match",
			host:    "example.com",
			noProxy: "other.com,another.com",
			want:    false,
		},
		{
			name:    "domain suffix match",
			host:    "sub.example.com",
			noProxy: ".example.com",
			want:    true,
		},
		{
			name:    "domain match",
			host:    "sub.example.com",
			noProxy: "example.com",
			want:    true,
		},
		{
			name:    "localhost",
			host:    "localhost",
			noProxy: "localhost,127.0.0.1",
			want:    true,
		},
		{
			name:    "with port",
			host:    "example.com:8080",
			noProxy: "example.com",
			want:    true,
		},
		{
			name:    "empty no_proxy",
			host:    "example.com",
			noProxy: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldBypassProxy(tt.host, tt.noProxy)
			if got != tt.want {
				t.Errorf("shouldBypassProxy(%s, %s) = %v, want %v", tt.host, tt.noProxy, got, tt.want)
			}
		})
	}
}

func TestMaskURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "no credentials",
			url:  "https://example.com/path",
			want: "https://example.com/path",
		},
		{
			name: "with username",
			url:  "https://user@example.com/path",
			want: "https://%2A%2A%2A:%2A%2A%2A@example.com/path",
		},
		{
			name: "with username and password",
			url:  "https://user:pass@example.com/path",
			want: "https://%2A%2A%2A:%2A%2A%2A@example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskURL(tt.url)
			if got != tt.want {
				t.Errorf("maskURL(%s) = %s, want %s", tt.url, got, tt.want)
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	resp := &http.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Body:       io.NopCloser(strings.NewReader("Resource not found")),
	}

	err := NewHTTPError(resp)
	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatal("NewHTTPError() did not return *HTTPError")
	}

	if httpErr.StatusCode != 404 {
		t.Errorf("HTTPError.StatusCode = %d, want 404", httpErr.StatusCode)
	}

	if !strings.Contains(httpErr.Error(), "404") {
		t.Errorf("HTTPError.Error() should contain '404'")
	}

	if !strings.Contains(httpErr.Error(), "Resource not found") {
		t.Errorf("HTTPError.Error() should contain error body")
	}
}
