package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/codeownersnet/atlas/internal/auth"
	"github.com/rs/zerolog"
	"golang.org/x/net/proxy"
)

const (
	defaultTimeout       = 30 * time.Second
	defaultMaxRetries    = 3
	defaultRetryDelay    = 1 * time.Second
	defaultMaxRetryDelay = 10 * time.Second
)

// Client is an HTTP client with retry logic, authentication, and logging
type Client struct {
	httpClient    *http.Client
	auth          auth.Provider
	baseURL       string
	customHeaders map[string]string
	logger        *zerolog.Logger
	maxRetries    int
	retryDelay    time.Duration
}

// Config holds the configuration for creating a new client
type Config struct {
	BaseURL       string
	Auth          auth.Provider
	CustomHeaders map[string]string
	Logger        *zerolog.Logger
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	SSLVerify     bool
	HTTPProxy     string
	HTTPSProxy    string
	SOCKSProxy    string
	NoProxy       string
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(cfg *Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if cfg.Auth == nil {
		return nil, fmt.Errorf("authentication provider is required")
	}

	// Set defaults
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}

	retryDelay := cfg.RetryDelay
	if retryDelay == 0 {
		retryDelay = defaultRetryDelay
	}

	// Create HTTP transport with proxy support
	transport, err := createTransport(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return &Client{
		httpClient:    httpClient,
		auth:          cfg.Auth,
		baseURL:       strings.TrimRight(cfg.BaseURL, "/"),
		customHeaders: cfg.CustomHeaders,
		logger:        cfg.Logger,
		maxRetries:    maxRetries,
		retryDelay:    retryDelay,
	}, nil
}

// createTransport creates an HTTP transport with proxy and SSL configuration
func createTransport(cfg *Config) (http.RoundTripper, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.SSLVerify,
		},
	}

	// Configure proxy
	if err := configureProxy(transport, cfg); err != nil {
		return nil, err
	}

	return transport, nil
}

// configureProxy configures proxy settings for the transport
func configureProxy(transport *http.Transport, cfg *Config) error {
	// SOCKS proxy takes precedence
	if cfg.SOCKSProxy != "" {
		socksURL, err := url.Parse(cfg.SOCKSProxy)
		if err != nil {
			return fmt.Errorf("invalid SOCKS proxy URL: %w", err)
		}

		var auth *proxy.Auth
		if socksURL.User != nil {
			password, _ := socksURL.User.Password()
			auth = &proxy.Auth{
				User:     socksURL.User.Username(),
				Password: password,
			}
		}

		dialer, err := proxy.SOCKS5("tcp", socksURL.Host, auth, proxy.Direct)
		if err != nil {
			return fmt.Errorf("failed to create SOCKS proxy dialer: %w", err)
		}

		transport.Dial = dialer.Dial
		return nil
	}

	// HTTP/HTTPS proxy
	if cfg.HTTPProxy != "" || cfg.HTTPSProxy != "" {
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			// Check NO_PROXY
			if cfg.NoProxy != "" && shouldBypassProxy(req.URL.Host, cfg.NoProxy) {
				return nil, nil
			}

			// Use HTTPS proxy for HTTPS requests, HTTP proxy otherwise
			var proxyURL string
			if req.URL.Scheme == "https" && cfg.HTTPSProxy != "" {
				proxyURL = cfg.HTTPSProxy
			} else if cfg.HTTPProxy != "" {
				proxyURL = cfg.HTTPProxy
			}

			if proxyURL == "" {
				return nil, nil
			}

			return url.Parse(proxyURL)
		}
	}

	return nil
}

// shouldBypassProxy checks if the host should bypass the proxy based on NO_PROXY
func shouldBypassProxy(host, noProxy string) bool {
	if noProxy == "" {
		return false
	}

	// Remove port from host if present
	if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	entries := strings.Split(noProxy, ",")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Exact match
		if host == entry {
			return true
		}

		// Suffix match for domain wildcards (e.g., .example.com)
		if strings.HasPrefix(entry, ".") && strings.HasSuffix(host, entry) {
			return true
		}

		// Domain match (example.com matches sub.example.com)
		if strings.HasSuffix(host, "."+entry) {
			return true
		}
	}

	return false
}

// Do performs an HTTP request with retry logic
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying with exponential backoff
			delay := c.retryDelay * time.Duration(1<<uint(attempt-1))
			if delay > defaultMaxRetryDelay {
				delay = defaultMaxRetryDelay
			}

			c.logDebug("retrying request after delay", map[string]interface{}{
				"attempt": attempt,
				"delay":   delay.String(),
				"method":  method,
				"path":    path,
			})

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := c.doRequest(ctx, method, path, body)
		if err != nil {
			lastErr = err
			c.logDebug("request failed", map[string]interface{}{
				"attempt": attempt,
				"error":   err.Error(),
				"method":  method,
				"path":    path,
			})
			continue
		}

		// Check if we should retry based on status code
		if c.shouldRetry(resp.StatusCode) && attempt < c.maxRetries {
			resp.Body.Close()
			lastErr = fmt.Errorf("received status code %d", resp.StatusCode)
			c.logDebug("retrying due to status code", map[string]interface{}{
				"attempt":     attempt,
				"status_code": resp.StatusCode,
				"method":      method,
				"path":        path,
			})
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Build full URL
	fullURL := c.baseURL + path

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply authentication
	if err := c.auth.Apply(req); err != nil {
		return nil, fmt.Errorf("failed to apply authentication: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Apply custom headers
	for key, value := range c.customHeaders {
		req.Header.Set(key, value)
	}

	// Log request (with sensitive data masked)
	c.logRequest(req)

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Log response
	c.logResponse(resp)

	return resp, nil
}

// shouldRetry determines if a request should be retried based on status code
func (c *Client) shouldRetry(statusCode int) bool {
	// Retry on server errors and rate limiting
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout ||
		(statusCode >= 500 && statusCode < 600)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, path, bytes.NewReader(body))
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body []byte) (*http.Response, error) {
	return c.Do(ctx, http.MethodPut, path, bytes.NewReader(body))
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.Do(ctx, http.MethodDelete, path, nil)
}

// Logging helpers

func (c *Client) logRequest(req *http.Request) {
	if c.logger == nil {
		return
	}

	fields := map[string]interface{}{
		"method": req.Method,
		"url":    maskURL(req.URL.String()),
		"auth":   c.auth.Mask(),
	}

	// Add custom headers (masked)
	if len(c.customHeaders) > 0 {
		maskedHeaders := make(map[string]string)
		for k := range c.customHeaders {
			maskedHeaders[k] = "***"
		}
		fields["custom_headers"] = maskedHeaders
	}

	c.logDebug("sending request", fields)
}

func (c *Client) logResponse(resp *http.Response) {
	if c.logger == nil {
		return
	}

	c.logDebug("received response", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	})
}

func (c *Client) logDebug(msg string, fields map[string]interface{}) {
	if c.logger == nil {
		return
	}

	event := c.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// maskURL masks sensitive information in URLs (credentials)
func maskURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if u.User != nil {
		// Mask username and password
		username := u.User.Username()
		if username != "" {
			u.User = url.UserPassword("***", "***")
		}
	}

	return u.String()
}

// Error types

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s - %s", e.StatusCode, e.Status, e.Body)
}

// NewHTTPError creates a new HTTP error from a response
func NewHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       string(body),
	}
}
