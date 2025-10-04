package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockAuth is a mock authentication provider for testing
type mockAuth struct{}

func (m *mockAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer test-token")
	return nil
}

func (m *mockAuth) Type() string {
	return "mock"
}

func (m *mockAuth) Mask() string {
	return "mock:***"
}

func TestDetectDeploymentType(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected DeploymentType
	}{
		{
			name:     "Cloud URL",
			url:      "https://mycompany.atlassian.net",
			expected: DeploymentCloud,
		},
		{
			name:     "Server URL",
			url:      "https://confluence.mycompany.com",
			expected: DeploymentServer,
		},
		{
			name:     "Server URL with port",
			url:      "http://localhost:8090",
			expected: DeploymentServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectDeploymentType(tt.url)
			if result != tt.expected {
				t.Errorf("detectDeploymentType(%s) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				BaseURL:   "https://confluence.example.com",
				Auth:      &mockAuth{},
				SSLVerify: true,
			},
			wantErr: false,
		},
		{
			name: "Missing base URL",
			config: &Config{
				Auth:      &mockAuth{},
				SSLVerify: true,
			},
			wantErr: true,
		},
		{
			name: "Missing auth",
			config: &Config{
				BaseURL:   "https://confluence.example.com",
				SSLVerify: true,
			},
			wantErr: true,
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
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestGetContent(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/content/123456" {
			t.Errorf("Expected path /rest/api/content/123456, got %s", r.URL.Path)
		}

		content := Content{
			ID:    "123456",
			Type:  ContentTypePage,
			Title: "Test Page",
			Space: &Space{
				Key:  "TEST",
				Name: "Test Space",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(content)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL:   server.URL,
		Auth:      &mockAuth{},
		SSLVerify: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	content, err := client.GetContent(context.Background(), "123456", nil)
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}

	if content.ID != "123456" {
		t.Errorf("Expected content ID 123456, got %s", content.ID)
	}
	if content.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got %s", content.Title)
	}
}

func TestSearchCQL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received request: %s %s", r.Method, r.URL.Path)
		t.Logf("Query: %s", r.URL.RawQuery)

		// Accept any path - the query parameters will be in r.URL.RawQuery
		result := SearchResult{
			Size:  1,
			Limit: 25,
			Start: 0,
			Results: []Content{
				{
					ID:    "123456",
					Type:  ContentTypePage,
					Title: "Test Page",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL:   server.URL,
		Auth:      &mockAuth{},
		SSLVerify: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err := client.SearchCQL(context.Background(), "type=page and space=TEST", nil)
	if err != nil {
		t.Fatalf("SearchCQL() error = %v", err)
	}

	if result.Size != 1 {
		t.Errorf("Expected size 1, got %d", result.Size)
	}
	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}
}

func TestGetSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/space" {
			t.Errorf("Expected path /rest/api/space, got %s", r.URL.Path)
		}

		response := struct {
			Results []Space `json:"results"`
		}{
			Results: []Space{
				{
					Key:  "TEST",
					Name: "Test Space",
					Type: "global",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL:   server.URL,
		Auth:      &mockAuth{},
		SSLVerify: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	spaces, err := client.GetSpaces(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetSpaces() error = %v", err)
	}

	if len(spaces) != 1 {
		t.Errorf("Expected 1 space, got %d", len(spaces))
	}
	if spaces[0].Key != "TEST" {
		t.Errorf("Expected space key TEST, got %s", spaces[0].Key)
	}
}

func TestErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		errResp := ErrorResponse{
			StatusCode: 404,
			Message:    "Content not found",
		}
		json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL:   server.URL,
		Auth:      &mockAuth{},
		SSLVerify: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetContent(context.Background(), "invalid", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestIsCQL(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{
			name:  "CQL with type",
			query: "type=page",
			want:  true,
		},
		{
			name:  "CQL with AND",
			query: "type=page AND space=TEST",
			want:  true,
		},
		{
			name:  "Simple text",
			query: "hello world",
			want:  false,
		},
		{
			name:  "Text with tilde",
			query: "text~test",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCQL(tt.query)
			if got != tt.want {
				t.Errorf("isCQL(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
