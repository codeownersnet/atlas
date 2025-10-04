package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/codeownersnet/atlas/internal/auth"
	"github.com/codeownersnet/atlas/internal/client"
)

const (
	// API paths
	apiVersion2  = "/rest/api/2"
	apiVersion3  = "/rest/api/3"
	agileVersion = "/rest/agile/1.0"
)

// Client is a Jira API client
type Client struct {
	httpClient     *client.Client
	baseURL        string
	deploymentType DeploymentType
}

// Config holds the configuration for creating a Jira client
type Config struct {
	BaseURL       string
	Auth          auth.Provider
	CustomHeaders map[string]string
	SSLVerify     bool
	HTTPProxy     string
	HTTPSProxy    string
	SOCKSProxy    string
	NoProxy       string
}

// NewClient creates a new Jira client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if cfg.Auth == nil {
		return nil, fmt.Errorf("auth provider is required")
	}

	// Detect deployment type from URL
	deploymentType := detectDeploymentType(cfg.BaseURL)

	// Create HTTP client
	httpClient, err := client.NewClient(&client.Config{
		BaseURL:       cfg.BaseURL,
		Auth:          cfg.Auth,
		CustomHeaders: cfg.CustomHeaders,
		SSLVerify:     cfg.SSLVerify,
		HTTPProxy:     cfg.HTTPProxy,
		HTTPSProxy:    cfg.HTTPSProxy,
		SOCKSProxy:    cfg.SOCKSProxy,
		NoProxy:       cfg.NoProxy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Client{
		httpClient:     httpClient,
		baseURL:        strings.TrimRight(cfg.BaseURL, "/"),
		deploymentType: deploymentType,
	}, nil
}

// detectDeploymentType detects if the Jira instance is Cloud or Server/DC
func detectDeploymentType(baseURL string) DeploymentType {
	if strings.Contains(baseURL, ".atlassian.net") {
		return DeploymentCloud
	}
	return DeploymentServer
}

// IsCloud returns true if the Jira instance is Cloud
func (c *Client) IsCloud() bool {
	return c.deploymentType == DeploymentCloud
}

// IsServer returns true if the Jira instance is Server/Data Center
func (c *Client) IsServer() bool {
	return c.deploymentType == DeploymentServer
}

// GetDeploymentType returns the deployment type
func (c *Client) GetDeploymentType() DeploymentType {
	return c.deploymentType
}

// doRequest performs an HTTP request and decodes the response
func (c *Client) doRequest(ctx context.Context, method, path string, body []byte, result interface{}) error {
	var resp *http.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = c.httpClient.Get(ctx, path)
	case http.MethodPost:
		resp, err = c.httpClient.Post(ctx, path, body)
	case http.MethodPut:
		resp, err = c.httpClient.Put(ctx, path, body)
	case http.MethodDelete:
		resp, err = c.httpClient.Delete(ctx, path)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return c.parseError(resp.StatusCode, respBody)
	}

	// Decode response if result is provided
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// parseError parses an error response from Jira
func (c *Client) parseError(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// If we can't parse the error, return the raw body
		return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
	}

	// Build error message
	var messages []string
	messages = append(messages, errResp.ErrorMessages...)
	for field, msg := range errResp.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, msg))
	}

	if len(messages) == 0 {
		return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
	}

	return fmt.Errorf("HTTP %d: %s", statusCode, strings.Join(messages, "; "))
}

// buildURL builds a full URL with query parameters
func buildURL(base string, params map[string]string) string {
	if len(params) == 0 {
		return base
	}

	values := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			values = append(values, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
		}
	}

	if len(values) == 0 {
		return base
	}

	return base + "?" + strings.Join(values, "&")
}

// getAPIPath returns the appropriate API path based on deployment type
func (c *Client) getAPIPath() string {
	// Both Cloud and Server/DC use API v2 for most operations
	return apiVersion2
}

// getSearchAPIPath returns the appropriate API path for search operations
// Cloud uses /rest/api/3/search/jql (v2 was deprecated and removed)
// Server/DC uses /rest/api/2/search
func (c *Client) getSearchAPIPath() string {
	if c.IsCloud() {
		return apiVersion3 + "/search/jql"
	}
	return apiVersion2 + "/search"
}

// getAgileAPIPath returns the agile API path
func (c *Client) getAgileAPIPath() string {
	return agileVersion
}
