package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds the complete application configuration
type Config struct {
	Jira       *JiraConfig
	Confluence *ConfluenceConfig
	Opsgenie   *OpsgenieConfig
	Server     *ServerConfig
	Security   *SecurityConfig
	Logging    *LoggingConfig
	Proxy      *ProxyConfig
}

// JiraConfig holds Jira-specific configuration
type JiraConfig struct {
	URL              string
	Username         string
	APIToken         string
	PersonalToken    string
	OAuthAccessToken string
	OAuthCloudID     string
	SSLVerify        bool
	ProjectsFilter   []string
	CustomHeaders    map[string]string
	HTTPProxy        string
	HTTPSProxy       string
	SOCKSProxy       string
	NoProxy          string
	AuthMethod       AuthMethod
}

// ConfluenceConfig holds Confluence-specific configuration
type ConfluenceConfig struct {
	URL              string
	Username         string
	APIToken         string
	PersonalToken    string
	OAuthAccessToken string
	OAuthCloudID     string
	SSLVerify        bool
	SpacesFilter     []string
	CustomHeaders    map[string]string
	HTTPProxy        string
	HTTPSProxy       string
	SOCKSProxy       string
	NoProxy          string
	AuthMethod       AuthMethod
}

// OpsgenieConfig holds Opsgenie-specific configuration
type OpsgenieConfig struct {
	URL           string
	APIKey        string
	SSLVerify     bool
	HTTPProxy     string
	HTTPSProxy    string
	SOCKSProxy    string
	NoProxy       string
	CustomHeaders map[string]string
}

// ServerConfig holds server transport configuration
type ServerConfig struct {
	Transport string // stdio (only supported transport)
	Port      int    // Reserved for future use
	Host      string // Reserved for future use
}

// SecurityConfig holds security and access control settings
type SecurityConfig struct {
	ReadOnlyMode bool
	EnabledTools []string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Verbose     bool
	VeryVerbose bool
	LogToStdout bool
}

// ProxyConfig holds global proxy configuration
type ProxyConfig struct {
	HTTPProxy  string
	HTTPSProxy string
	SOCKSProxy string
	NoProxy    string
}

// AuthMethod represents the authentication method to use
type AuthMethod int

const (
	AuthMethodUnknown AuthMethod = iota
	AuthMethodBasic              // Username + API Token (Cloud)
	AuthMethodPAT                // Personal Access Token (Server/DC)
	AuthMethodOAuth              // Bearer Token (BYO - Bring Your Own)
)

func (a AuthMethod) String() string {
	switch a {
	case AuthMethodBasic:
		return "basic"
	case AuthMethodPAT:
		return "pat"
	case AuthMethodOAuth:
		return "oauth"
	default:
		return "unknown"
	}
}

// Load loads configuration from environment variables, .env file, and CLI flags
func Load(configFile ...string) (*Config, error) {
	// Load .env file if it exists (ignore errors if file doesn't exist)
	// Use Overload to allow config file to override shell environment variables
	if len(configFile) > 0 && configFile[0] != "" {
		// Load specified config file and override existing env vars
		if err := godotenv.Overload(configFile[0]); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile[0], err)
		}
	} else {
		// Load default .env file and override existing env vars (ignore errors if file doesn't exist)
		_ = godotenv.Overload()
	}

	// Initialize viper
	viper.AutomaticEnv()

	cfg := &Config{
		Jira:       loadJiraConfig(),
		Confluence: loadConfluenceConfig(),
		Opsgenie:   loadOpsgenieConfig(),
		Server:     loadServerConfig(),
		Security:   loadSecurityConfig(),
		Logging:    loadLoggingConfig(),
		Proxy:      loadProxyConfig(),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadJiraConfig loads Jira-specific configuration
func loadJiraConfig() *JiraConfig {
	cfg := &JiraConfig{
		URL:              getEnv("JIRA_URL", ""),
		Username:         getEnv("JIRA_USERNAME", ""),
		APIToken:         getEnv("JIRA_API_TOKEN", ""),
		PersonalToken:    getEnv("JIRA_PERSONAL_TOKEN", ""),
		OAuthAccessToken: getEnv("ATLASSIAN_OAUTH_ACCESS_TOKEN", ""),
		OAuthCloudID:     getEnv("ATLASSIAN_OAUTH_CLOUD_ID", ""),
		SSLVerify:        getEnvBool("JIRA_SSL_VERIFY", true),
		ProjectsFilter:   getEnvList("JIRA_PROJECTS_FILTER", []string{}),
		CustomHeaders:    parseCustomHeaders(getEnv("JIRA_CUSTOM_HEADERS", "")),
		HTTPProxy:        getEnv("JIRA_HTTP_PROXY", ""),
		HTTPSProxy:       getEnv("JIRA_HTTPS_PROXY", ""),
		SOCKSProxy:       getEnv("JIRA_SOCKS_PROXY", ""),
		NoProxy:          getEnv("JIRA_NO_PROXY", ""),
	}

	// Detect auth method
	cfg.AuthMethod = detectAuthMethod(cfg.Username, cfg.APIToken, cfg.PersonalToken, cfg.OAuthAccessToken)

	return cfg
}

// loadConfluenceConfig loads Confluence-specific configuration
func loadConfluenceConfig() *ConfluenceConfig {
	cfg := &ConfluenceConfig{
		URL:              getEnv("CONFLUENCE_URL", ""),
		Username:         getEnv("CONFLUENCE_USERNAME", ""),
		APIToken:         getEnv("CONFLUENCE_API_TOKEN", ""),
		PersonalToken:    getEnv("CONFLUENCE_PERSONAL_TOKEN", ""),
		OAuthAccessToken: getEnv("ATLASSIAN_OAUTH_ACCESS_TOKEN", ""),
		OAuthCloudID:     getEnv("ATLASSIAN_OAUTH_CLOUD_ID", ""),
		SSLVerify:        getEnvBool("CONFLUENCE_SSL_VERIFY", true),
		SpacesFilter:     getEnvList("CONFLUENCE_SPACES_FILTER", []string{}),
		CustomHeaders:    parseCustomHeaders(getEnv("CONFLUENCE_CUSTOM_HEADERS", "")),
		HTTPProxy:        getEnv("CONFLUENCE_HTTP_PROXY", ""),
		HTTPSProxy:       getEnv("CONFLUENCE_HTTPS_PROXY", ""),
		SOCKSProxy:       getEnv("CONFLUENCE_SOCKS_PROXY", ""),
		NoProxy:          getEnv("CONFLUENCE_NO_PROXY", ""),
	}

	// Detect auth method
	cfg.AuthMethod = detectAuthMethod(cfg.Username, cfg.APIToken, cfg.PersonalToken, cfg.OAuthAccessToken)

	return cfg
}

// loadOpsgenieConfig loads Opsgenie-specific configuration
func loadOpsgenieConfig() *OpsgenieConfig {
	cfg := &OpsgenieConfig{
		URL:           getEnv("OPSGENIE_URL", ""),
		APIKey:        getEnv("OPSGENIE_API_KEY", ""),
		SSLVerify:     getEnvBool("OPSGENIE_SSL_VERIFY", true),
		CustomHeaders: parseCustomHeaders(getEnv("OPSGENIE_CUSTOM_HEADERS", "")),
		HTTPProxy:     getEnv("OPSGENIE_HTTP_PROXY", ""),
		HTTPSProxy:    getEnv("OPSGENIE_HTTPS_PROXY", ""),
		SOCKSProxy:    getEnv("OPSGENIE_SOCKS_PROXY", ""),
		NoProxy:       getEnv("OPSGENIE_NO_PROXY", ""),
	}

	return cfg
}

// loadServerConfig loads server transport configuration
func loadServerConfig() *ServerConfig {
	return &ServerConfig{
		Transport: getEnv("TRANSPORT", "stdio"),
		Port:      getEnvInt("PORT", 8000),
		Host:      getEnv("HOST", "0.0.0.0"),
	}
}

// loadSecurityConfig loads security and access control settings
func loadSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ReadOnlyMode: getEnvBool("READ_ONLY_MODE", false),
		EnabledTools: getEnvList("ENABLED_TOOLS", []string{}),
	}
}

// loadLoggingConfig loads logging configuration
func loadLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Verbose:     getEnvBool("MCP_VERBOSE", false),
		VeryVerbose: getEnvBool("MCP_VERY_VERBOSE", false),
		LogToStdout: getEnvBool("MCP_LOGGING_STDOUT", false),
	}
}

// loadProxyConfig loads global proxy configuration
func loadProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		HTTPProxy:  getEnv("HTTP_PROXY", ""),
		HTTPSProxy: getEnv("HTTPS_PROXY", ""),
		SOCKSProxy: getEnv("SOCKS_PROXY", ""),
		NoProxy:    getEnv("NO_PROXY", ""),
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// At least one service must be configured
	jiraConfigured := c.Jira != nil && c.Jira.URL != ""
	confluenceConfigured := c.Confluence != nil && c.Confluence.URL != ""
	opsgenieConfigured := c.Opsgenie != nil && c.Opsgenie.APIKey != ""

	if !jiraConfigured && !confluenceConfigured && !opsgenieConfigured {
		return fmt.Errorf("at least one service (Jira, Confluence, or Opsgenie) must be configured")
	}

	// Validate Jira configuration if provided
	if jiraConfigured {
		if err := c.Jira.Validate(); err != nil {
			return fmt.Errorf("jira configuration: %w", err)
		}
	}

	// Validate Confluence configuration if provided
	if confluenceConfigured {
		if err := c.Confluence.Validate(); err != nil {
			return fmt.Errorf("confluence configuration: %w", err)
		}
	}

	// Validate Opsgenie configuration if provided
	if opsgenieConfigured {
		if err := c.Opsgenie.Validate(); err != nil {
			return fmt.Errorf("opsgenie configuration: %w", err)
		}
	}

	// Validate server configuration
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server configuration: %w", err)
	}

	return nil
}

// Validate validates Jira configuration
func (j *JiraConfig) Validate() error {
	if j.URL == "" {
		return fmt.Errorf("JIRA_URL is required")
	}

	// Validate URL format
	if _, err := url.Parse(j.URL); err != nil {
		return fmt.Errorf("invalid JIRA_URL: %w", err)
	}

	// Validate auth configuration based on method
	switch j.AuthMethod {
	case AuthMethodBasic:
		if j.Username == "" || j.APIToken == "" {
			return fmt.Errorf("basic auth requires both JIRA_USERNAME and JIRA_API_TOKEN")
		}
	case AuthMethodPAT:
		if j.PersonalToken == "" {
			return fmt.Errorf("PAT auth requires JIRA_PERSONAL_TOKEN")
		}
	case AuthMethodOAuth:
		if j.OAuthAccessToken == "" {
			return fmt.Errorf("OAuth auth requires ATLASSIAN_OAUTH_ACCESS_TOKEN")
		}
	}

	return nil
}

// Validate validates Confluence configuration
func (c *ConfluenceConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("CONFLUENCE_URL is required")
	}

	// Validate URL format
	if _, err := url.Parse(c.URL); err != nil {
		return fmt.Errorf("invalid CONFLUENCE_URL: %w", err)
	}

	// Validate auth configuration based on method
	switch c.AuthMethod {
	case AuthMethodBasic:
		if c.Username == "" || c.APIToken == "" {
			return fmt.Errorf("basic auth requires both CONFLUENCE_USERNAME and CONFLUENCE_API_TOKEN")
		}
	case AuthMethodPAT:
		if c.PersonalToken == "" {
			return fmt.Errorf("PAT auth requires CONFLUENCE_PERSONAL_TOKEN")
		}
	case AuthMethodOAuth:
		if c.OAuthAccessToken == "" {
			return fmt.Errorf("OAuth auth requires ATLASSIAN_OAUTH_ACCESS_TOKEN")
		}
	}

	return nil
}

// Validate validates Opsgenie configuration
func (o *OpsgenieConfig) Validate() error {
	if o.APIKey == "" {
		return fmt.Errorf("OPSGENIE_API_KEY is required")
	}

	// URL is optional - will default to Opsgenie's public API if not provided
	if o.URL != "" {
		// Validate URL format if provided
		if _, err := url.Parse(o.URL); err != nil {
			return fmt.Errorf("invalid OPSGENIE_URL: %w", err)
		}
	}

	return nil
}

// Validate validates server configuration
func (s *ServerConfig) Validate() error {
	// Only stdio transport is supported
	if s.Transport != "" && s.Transport != "stdio" {
		// Don't fail validation, just default to stdio
		s.Transport = "stdio"
	}

	if s.Transport == "" {
		s.Transport = "stdio"
	}

	return nil
}

// IsJiraConfigured returns true if Jira is configured
func (c *Config) IsJiraConfigured() bool {
	return c.Jira != nil && c.Jira.URL != ""
}

// IsConfluenceConfigured returns true if Confluence is configured
func (c *Config) IsConfluenceConfigured() bool {
	return c.Confluence != nil && c.Confluence.URL != ""
}

// IsOpsgenieConfigured returns true if Opsgenie is configured
func (c *Config) IsOpsgenieConfigured() bool {
	return c.Opsgenie != nil && c.Opsgenie.APIKey != ""
}

// detectAuthMethod detects the authentication method based on provided credentials
func detectAuthMethod(username, apiToken, personalToken, oauthAccessToken string) AuthMethod {
	if oauthAccessToken != "" {
		return AuthMethodOAuth
	}
	if personalToken != "" {
		return AuthMethodPAT
	}
	if username != "" && apiToken != "" {
		return AuthMethodBasic
	}
	return AuthMethodUnknown
}

// parseCustomHeaders parses comma-separated key=value pairs into a map
func parseCustomHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				headers[key] = value
			}
		}
	}

	return headers
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

func getEnvList(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
