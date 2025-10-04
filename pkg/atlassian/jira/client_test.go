package jira

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
			url:      "https://jira.mycompany.com",
			expected: DeploymentServer,
		},
		{
			name:     "Server URL with port",
			url:      "http://localhost:8080",
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

func TestGetSearchAPIPath(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		expectedPath string
	}{
		{
			name:         "Cloud instance uses v3 search/jql",
			baseURL:      "https://mycompany.atlassian.net",
			expectedPath: "/rest/api/3/search/jql",
		},
		{
			name:         "Server instance uses v2 search",
			baseURL:      "https://jira.mycompany.com",
			expectedPath: "/rest/api/2/search",
		},
		{
			name:         "Server with port uses v2 search",
			baseURL:      "http://localhost:8080",
			expectedPath: "/rest/api/2/search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(&Config{
				BaseURL:   tt.baseURL,
				Auth:      &mockAuth{},
				SSLVerify: true,
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result := client.getSearchAPIPath()
			if result != tt.expectedPath {
				t.Errorf("getSearchAPIPath() = %s, want %s", result, tt.expectedPath)
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
				BaseURL:   "https://jira.example.com",
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
				BaseURL:   "https://jira.example.com",
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

func TestGetIssue(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/issue/TEST-123" {
			t.Errorf("Expected path /rest/api/2/issue/TEST-123, got %s", r.URL.Path)
		}

		issue := Issue{
			ID:  "10001",
			Key: "TEST-123",
			Fields: IssueFields{
				Summary: "Test issue",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
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

	issue, err := client.GetIssue(context.Background(), "TEST-123", nil)
	if err != nil {
		t.Fatalf("GetIssue() error = %v", err)
	}

	if issue.Key != "TEST-123" {
		t.Errorf("Expected issue key TEST-123, got %s", issue.Key)
	}
	if issue.Fields.Summary != "Test issue" {
		t.Errorf("Expected summary 'Test issue', got %s", issue.Fields.Summary)
	}
}

func TestSearchIssues(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		expectedPath string
	}{
		{
			name:         "Server deployment",
			baseURL:      "https://jira.mycompany.com",
			expectedPath: "/rest/api/2/search",
		},
		{
			name:         "Cloud deployment",
			baseURL:      "https://mycompany.atlassian.net",
			expectedPath: "/rest/api/3/search/jql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedPath := ""
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path

				result := SearchResult{
					Total:      1,
					MaxResults: 50,
					StartAt:    0,
					Issues: []Issue{
						{
							ID:  "10001",
							Key: "TEST-123",
							Fields: IssueFields{
								Summary: "Test issue",
							},
						},
					},
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(result)
			}))
			defer server.Close()

			// Create client with the deployment-specific baseURL to detect type correctly
			client, err := NewClient(&Config{
				BaseURL:   tt.baseURL,
				Auth:      &mockAuth{},
				SSLVerify: true,
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Override the httpClient with one pointing to test server
			// while preserving deployment type in the Jira client
			testClient, err := NewClient(&Config{
				BaseURL:   server.URL,
				Auth:      &mockAuth{},
				SSLVerify: true,
			})
			if err != nil {
				t.Fatalf("Failed to create test client: %v", err)
			}
			client.httpClient = testClient.httpClient
			client.baseURL = server.URL

			result, err := client.SearchIssues(context.Background(), "project = TEST", nil)
			if err != nil {
				t.Fatalf("SearchIssues() error = %v", err)
			}

			// Verify the correct endpoint was called
			if receivedPath != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, receivedPath)
			}

			if result.Total != 1 {
				t.Errorf("Expected total 1, got %d", result.Total)
			}
			if len(result.Issues) != 1 {
				t.Errorf("Expected 1 issue, got %d", len(result.Issues))
			}
		})
	}
}

func TestGetAllProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/project" {
			t.Errorf("Expected path /rest/api/2/project, got %s", r.URL.Path)
		}

		projects := []Project{
			{
				ID:   "10000",
				Key:  "TEST",
				Name: "Test Project",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
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

	projects, err := client.GetAllProjects(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetAllProjects() error = %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
	if projects[0].Key != "TEST" {
		t.Errorf("Expected project key TEST, got %s", projects[0].Key)
	}
}

func TestErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		errResp := ErrorResponse{
			ErrorMessages: []string{"Issue does not exist"},
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

	_, err = client.GetIssue(context.Background(), "INVALID-123", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
