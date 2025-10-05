package config

import (
	"os"
	"testing"
)

func TestDetectAuthMethod(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		apiToken         string
		personalToken    string
		oauthAccessToken string
		want             AuthMethod
	}{
		{
			name:             "oauth auth",
			username:         "",
			apiToken:         "",
			personalToken:    "",
			oauthAccessToken: "oauth-token-123",
			want:             AuthMethodOAuth,
		},
		{
			name:             "oauth takes precedence",
			username:         "user@example.com",
			apiToken:         "token123",
			personalToken:    "pat123",
			oauthAccessToken: "oauth-token-123",
			want:             AuthMethodOAuth,
		},
		{
			name:             "basic auth",
			username:         "user@example.com",
			apiToken:         "token123",
			personalToken:    "",
			oauthAccessToken: "",
			want:             AuthMethodBasic,
		},
		{
			name:             "PAT auth",
			username:         "",
			apiToken:         "",
			personalToken:    "pat123",
			oauthAccessToken: "",
			want:             AuthMethodPAT,
		},
		{
			name:             "PAT takes precedence over basic",
			username:         "user@example.com",
			apiToken:         "token123",
			personalToken:    "pat123",
			oauthAccessToken: "",
			want:             AuthMethodPAT,
		},
		{
			name:             "no auth",
			username:         "",
			apiToken:         "",
			personalToken:    "",
			oauthAccessToken: "",
			want:             AuthMethodUnknown,
		},
		{
			name:             "incomplete basic auth - no token",
			username:         "user@example.com",
			apiToken:         "",
			personalToken:    "",
			oauthAccessToken: "",
			want:             AuthMethodUnknown,
		},
		{
			name:             "incomplete basic auth - no username",
			username:         "",
			apiToken:         "token123",
			personalToken:    "",
			oauthAccessToken: "",
			want:             AuthMethodUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectAuthMethod(tt.username, tt.apiToken, tt.personalToken, tt.oauthAccessToken)
			if got != tt.want {
				t.Errorf("detectAuthMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCustomHeaders(t *testing.T) {
	tests := []struct {
		name      string
		headerStr string
		want      map[string]string
	}{
		{
			name:      "empty string",
			headerStr: "",
			want:      map[string]string{},
		},
		{
			name:      "single header",
			headerStr: "X-Custom=value1",
			want: map[string]string{
				"X-Custom": "value1",
			},
		},
		{
			name:      "multiple headers",
			headerStr: "X-Custom=value1,X-Another=value2",
			want: map[string]string{
				"X-Custom":  "value1",
				"X-Another": "value2",
			},
		},
		{
			name:      "headers with spaces",
			headerStr: " X-Custom = value1 , X-Another = value2 ",
			want: map[string]string{
				"X-Custom":  "value1",
				"X-Another": "value2",
			},
		},
		{
			name:      "header with equals in value",
			headerStr: "X-Custom=value=with=equals",
			want: map[string]string{
				"X-Custom": "value=with=equals",
			},
		},
		{
			name:      "malformed header ignored",
			headerStr: "X-Valid=value1,InvalidHeader,X-Another=value2",
			want: map[string]string{
				"X-Valid":   "value1",
				"X-Another": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCustomHeaders(tt.headerStr)
			if len(got) != len(tt.want) {
				t.Errorf("parseCustomHeaders() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseCustomHeaders()[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestGetEnvList(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue []string
		want         []string
	}{
		{
			name:         "empty env returns default",
			envValue:     "",
			defaultValue: []string{"default1", "default2"},
			want:         []string{"default1", "default2"},
		},
		{
			name:         "single value",
			envValue:     "value1",
			defaultValue: []string{},
			want:         []string{"value1"},
		},
		{
			name:         "multiple values",
			envValue:     "value1,value2,value3",
			defaultValue: []string{},
			want:         []string{"value1", "value2", "value3"},
		},
		{
			name:         "values with spaces",
			envValue:     " value1 , value2 , value3 ",
			defaultValue: []string{},
			want:         []string{"value1", "value2", "value3"},
		},
		{
			name:         "empty items filtered",
			envValue:     "value1,,value2,  ,value3",
			defaultValue: []string{},
			want:         []string{"value1", "value2", "value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env var for this test
			if tt.envValue != "" {
				os.Setenv("TEST_LIST", tt.envValue)
				defer os.Unsetenv("TEST_LIST")
			}

			got := getEnvList("TEST_LIST", tt.defaultValue)
			if len(got) != len(tt.want) {
				t.Errorf("getEnvList() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("getEnvList()[%d] = %v, want %v", i, got[i], v)
				}
			}
		})
	}
}

func TestJiraConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *JiraConfig
		wantErr bool
	}{
		{
			name: "valid basic auth",
			config: &JiraConfig{
				URL:        "https://example.atlassian.net",
				Username:   "user@example.com",
				APIToken:   "token123",
				AuthMethod: AuthMethodBasic,
			},
			wantErr: false,
		},
		{
			name: "valid PAT auth",
			config: &JiraConfig{
				URL:           "https://jira.example.com",
				PersonalToken: "pat123",
				AuthMethod:    AuthMethodPAT,
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: &JiraConfig{
				Username:   "user@example.com",
				APIToken:   "token123",
				AuthMethod: AuthMethodBasic,
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			config: &JiraConfig{
				URL:        "not a valid url://",
				Username:   "user@example.com",
				APIToken:   "token123",
				AuthMethod: AuthMethodBasic,
			},
			wantErr: true,
		},
		{
			name: "unknown auth method",
			config: &JiraConfig{
				URL:        "https://example.atlassian.net",
				AuthMethod: AuthMethodUnknown,
			},
			wantErr: false,
		},
		{
			name: "basic auth missing username",
			config: &JiraConfig{
				URL:        "https://example.atlassian.net",
				APIToken:   "token123",
				AuthMethod: AuthMethodBasic,
			},
			wantErr: true,
		},
		{
			name: "basic auth missing token",
			config: &JiraConfig{
				URL:        "https://example.atlassian.net",
				Username:   "user@example.com",
				AuthMethod: AuthMethodBasic,
			},
			wantErr: true,
		},
		{
			name: "PAT auth missing token",
			config: &JiraConfig{
				URL:        "https://jira.example.com",
				AuthMethod: AuthMethodPAT,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("JiraConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServerConfigValidate(t *testing.T) {
	tests := []struct {
		name              string
		config            *ServerConfig
		wantErr           bool
		wantTransport     string
		checkTransport    bool
	}{
		{
			name: "valid stdio",
			config: &ServerConfig{
				Transport: "stdio",
				Port:      8000,
				Host:      "0.0.0.0",
			},
			wantErr:        false,
			wantTransport:  "stdio",
			checkTransport: true,
		},
		{
			name: "empty transport defaults to stdio",
			config: &ServerConfig{
				Transport: "",
				Port:      8080,
				Host:      "localhost",
			},
			wantErr:        false,
			wantTransport:  "stdio",
			checkTransport: true,
		},
		{
			name: "sse transport defaults to stdio",
			config: &ServerConfig{
				Transport: "sse",
				Port:      3000,
				Host:      "127.0.0.1",
			},
			wantErr:        false,
			wantTransport:  "stdio",
			checkTransport: true,
		},
		{
			name: "streamable-http transport defaults to stdio",
			config: &ServerConfig{
				Transport: "streamable-http",
				Port:      8000,
				Host:      "0.0.0.0",
			},
			wantErr:        false,
			wantTransport:  "stdio",
			checkTransport: true,
		},
		{
			name: "invalid transport defaults to stdio",
			config: &ServerConfig{
				Transport: "invalid",
				Port:      8000,
				Host:      "0.0.0.0",
			},
			wantErr:        false,
			wantTransport:  "stdio",
			checkTransport: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkTransport && tt.config.Transport != tt.wantTransport {
				t.Errorf("ServerConfig.Validate() transport = %v, want %v", tt.config.Transport, tt.wantTransport)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with Jira only",
			config: &Config{
				Jira: &JiraConfig{
					URL:        "https://example.atlassian.net",
					Username:   "user@example.com",
					APIToken:   "token123",
					AuthMethod: AuthMethodBasic,
				},
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: false,
		},
		{
			name: "valid config with Confluence only",
			config: &Config{
				Confluence: &ConfluenceConfig{
					URL:           "https://wiki.example.com",
					PersonalToken: "pat123",
					AuthMethod:    AuthMethodPAT,
				},
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: false,
		},
		{
			name: "valid config with both services",
			config: &Config{
				Jira: &JiraConfig{
					URL:        "https://example.atlassian.net",
					Username:   "user@example.com",
					APIToken:   "token123",
					AuthMethod: AuthMethodBasic,
				},
				Confluence: &ConfluenceConfig{
					URL:           "https://wiki.example.com",
					PersonalToken: "pat123",
					AuthMethod:    AuthMethodPAT,
				},
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: false,
		},
		{
			name: "no services configured",
			config: &Config{
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: true,
		},
		{
			name: "valid Jira config with unknown auth method",
			config: &Config{
				Jira: &JiraConfig{
					URL:        "https://example.atlassian.net",
					AuthMethod: AuthMethodUnknown,
				},
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: false,
		},
		{
			name: "server config with non-stdio transport defaults to stdio",
			config: &Config{
				Jira: &JiraConfig{
					URL:        "https://example.atlassian.net",
					Username:   "user@example.com",
					APIToken:   "token123",
					AuthMethod: AuthMethodBasic,
				},
				Server: &ServerConfig{
					Transport: "invalid",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: false, // No longer returns error, defaults to stdio
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthMethodString(t *testing.T) {
	tests := []struct {
		name   string
		method AuthMethod
		want   string
	}{
		{
			name:   "basic",
			method: AuthMethodBasic,
			want:   "basic",
		},
		{
			name:   "pat",
			method: AuthMethodPAT,
			want:   "pat",
		},
		{
			name:   "oauth",
			method: AuthMethodOAuth,
			want:   "oauth",
		},
		{
			name:   "unknown",
			method: AuthMethodUnknown,
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method.String()
			if got != tt.want {
				t.Errorf("AuthMethod.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpsgenieConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *OpsgenieConfig
		wantErr bool
	}{
		{
			name: "valid with URL",
			config: &OpsgenieConfig{
				URL:    "https://api.opsgenie.com",
				APIKey: "api-key-123",
			},
			wantErr: false,
		},
		{
			name: "valid without URL",
			config: &OpsgenieConfig{
				APIKey: "api-key-123",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: &OpsgenieConfig{
				URL: "https://api.opsgenie.com",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			config: &OpsgenieConfig{
				URL:    "not a valid url://",
				APIKey: "api-key-123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OpsgenieConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidateWithOpsgenie(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with Opsgenie only",
			config: &Config{
				Opsgenie: &OpsgenieConfig{
					APIKey: "api-key-123",
				},
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: false,
		},
		{
			name: "valid config with all services",
			config: &Config{
				Jira: &JiraConfig{
					URL:        "https://example.atlassian.net",
					Username:   "user@example.com",
					APIToken:   "token123",
					AuthMethod: AuthMethodBasic,
				},
				Confluence: &ConfluenceConfig{
					URL:           "https://wiki.example.com",
					PersonalToken: "pat123",
					AuthMethod:    AuthMethodPAT,
				},
				Opsgenie: &OpsgenieConfig{
					APIKey: "api-key-123",
				},
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: false,
		},
		{
			name: "invalid Opsgenie config",
			config: &Config{
				Opsgenie: &OpsgenieConfig{
					URL: "https://api.opsgenie.com",
					// Missing APIKey
				},
				Server: &ServerConfig{
					Transport: "stdio",
					Port:      8000,
					Host:      "0.0.0.0",
				},
				Security: &SecurityConfig{},
				Logging:  &LoggingConfig{},
				Proxy:    &ProxyConfig{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsOpsgenieConfigured(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name: "opsgenie configured",
			config: &Config{
				Opsgenie: &OpsgenieConfig{
					APIKey: "api-key-123",
				},
			},
			want: true,
		},
		{
			name: "opsgenie not configured - no API key",
			config: &Config{
				Opsgenie: &OpsgenieConfig{
					URL: "https://api.opsgenie.com",
				},
			},
			want: false,
		},
		{
			name: "opsgenie not configured - nil",
			config: &Config{
				Opsgenie: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsOpsgenieConfigured()
			if got != tt.want {
				t.Errorf("Config.IsOpsgenieConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}
