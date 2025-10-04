package confluence

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
	apiPath = "/rest/api"
)

// Client is a Confluence API client
type Client struct {
	httpClient     *client.Client
	baseURL        string
	deploymentType DeploymentType
}

// Config holds the configuration for creating a Confluence client
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

// NewClient creates a new Confluence client
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

// detectDeploymentType detects if the Confluence instance is Cloud or Server/DC
func detectDeploymentType(baseURL string) DeploymentType {
	if strings.Contains(baseURL, ".atlassian.net") {
		return DeploymentCloud
	}
	return DeploymentServer
}

// IsCloud returns true if the Confluence instance is Cloud
func (c *Client) IsCloud() bool {
	return c.deploymentType == DeploymentCloud
}

// IsServer returns true if the Confluence instance is Server/Data Center
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

// parseError parses an error response from Confluence
func (c *Client) parseError(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// If we can't parse the error, return the raw body
		return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
	}

	// Build error message
	if errResp.Message != "" {
		return fmt.Errorf("HTTP %d: %s", statusCode, errResp.Message)
	}

	if errResp.Reason != "" {
		return fmt.Errorf("HTTP %d: %s", statusCode, errResp.Reason)
	}

	return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
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

// getAPIPath returns the API path
func (c *Client) getAPIPath() string {
	return apiPath
}

// expandFields converts a list of expand fields to a comma-separated string
func expandFields(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, ",")
}
