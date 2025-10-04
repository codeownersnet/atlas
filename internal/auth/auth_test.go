package auth

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		apiToken string
		wantErr  bool
	}{
		{
			name:     "valid credentials",
			username: "user@example.com",
			apiToken: "token123",
			wantErr:  false,
		},
		{
			name:     "missing username",
			username: "",
			apiToken: "token123",
			wantErr:  true,
		},
		{
			name:     "missing token",
			username: "user@example.com",
			apiToken: "",
			wantErr:  true,
		},
		{
			name:     "both missing",
			username: "",
			apiToken: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewBasicAuth(tt.username, tt.apiToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBasicAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewBasicAuth() returned nil auth without error")
			}
		})
	}
}

func TestBasicAuthApply(t *testing.T) {
	auth, err := NewBasicAuth("user@example.com", "token123")
	if err != nil {
		t.Fatalf("Failed to create BasicAuth: %v", err)
	}

	tests := []struct {
		name    string
		req     *http.Request
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &http.Request{Header: make(http.Header)},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.Apply(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BasicAuth.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				authHeader := tt.req.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Basic ") {
					t.Errorf("BasicAuth.Apply() Authorization header = %v, want prefix 'Basic '", authHeader)
				}
			}
		})
	}
}

func TestBasicAuthType(t *testing.T) {
	auth, _ := NewBasicAuth("user@example.com", "token123")
	if got := auth.Type(); got != "basic" {
		t.Errorf("BasicAuth.Type() = %v, want 'basic'", got)
	}
}

func TestBasicAuthMask(t *testing.T) {
	auth, _ := NewBasicAuth("user@example.com", "token123456789")
	masked := auth.Mask()
	if !strings.Contains(masked, "user@example.com") {
		t.Errorf("BasicAuth.Mask() should contain username")
	}
	if strings.Contains(masked, "token123456789") {
		t.Errorf("BasicAuth.Mask() should not contain full token")
	}
	if !strings.Contains(masked, "****") {
		t.Errorf("BasicAuth.Mask() should contain masked characters")
	}
}

func TestNewPATAuth(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "pat_token_123",
			wantErr: false,
		},
		{
			name:    "missing token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewPATAuth(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPATAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewPATAuth() returned nil auth without error")
			}
		})
	}
}

func TestPATAuthApply(t *testing.T) {
	auth, err := NewPATAuth("pat_token_123")
	if err != nil {
		t.Fatalf("Failed to create PATAuth: %v", err)
	}

	tests := []struct {
		name    string
		req     *http.Request
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &http.Request{Header: make(http.Header)},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.Apply(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PATAuth.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				authHeader := tt.req.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					t.Errorf("PATAuth.Apply() Authorization header = %v, want prefix 'Bearer '", authHeader)
				}
			}
		})
	}
}

func TestPATAuthType(t *testing.T) {
	auth, _ := NewPATAuth("pat_token_123")
	if got := auth.Type(); got != "pat" {
		t.Errorf("PATAuth.Type() = %v, want 'pat'", got)
	}
}

func TestPATAuthMask(t *testing.T) {
	auth, _ := NewPATAuth("pat_token_123456789")
	masked := auth.Mask()
	if strings.Contains(masked, "pat_token_123456789") {
		t.Errorf("PATAuth.Mask() should not contain full token")
	}
	if !strings.Contains(masked, "****") {
		t.Errorf("PATAuth.Mask() should contain masked characters")
	}
}

func TestNewOAuthAuth(t *testing.T) {
	tests := []struct {
		name        string
		accessToken string
		cloudID     string
		wantErr     bool
	}{
		{
			name:        "valid with cloud ID",
			accessToken: "oauth_token_123",
			cloudID:     "cloud-123",
			wantErr:     false,
		},
		{
			name:        "valid without cloud ID",
			accessToken: "oauth_token_123",
			cloudID:     "",
			wantErr:     false,
		},
		{
			name:        "missing access token",
			accessToken: "",
			cloudID:     "cloud-123",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewOAuthAuth(tt.accessToken, tt.cloudID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOAuthAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewOAuthAuth() returned nil auth without error")
			}
		})
	}
}

func TestOAuthAuthApply(t *testing.T) {
	tests := []struct {
		name        string
		accessToken string
		cloudID     string
		req         *http.Request
		wantErr     bool
		wantCloudID bool
	}{
		{
			name:        "valid request with cloud ID",
			accessToken: "oauth_token_123",
			cloudID:     "cloud-123",
			req:         &http.Request{Header: make(http.Header)},
			wantErr:     false,
			wantCloudID: true,
		},
		{
			name:        "valid request without cloud ID",
			accessToken: "oauth_token_123",
			cloudID:     "",
			req:         &http.Request{Header: make(http.Header)},
			wantErr:     false,
			wantCloudID: false,
		},
		{
			name:        "nil request",
			accessToken: "oauth_token_123",
			cloudID:     "",
			req:         nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, _ := NewOAuthAuth(tt.accessToken, tt.cloudID)
			err := auth.Apply(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("OAuthAuth.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				authHeader := tt.req.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					t.Errorf("OAuthAuth.Apply() Authorization header = %v, want prefix 'Bearer '", authHeader)
				}
				cloudIDHeader := tt.req.Header.Get("X-Atlassian-Cloud-Id")
				if tt.wantCloudID && cloudIDHeader == "" {
					t.Error("OAuthAuth.Apply() should set X-Atlassian-Cloud-Id header")
				}
				if !tt.wantCloudID && cloudIDHeader != "" {
					t.Error("OAuthAuth.Apply() should not set X-Atlassian-Cloud-Id header when cloudID is empty")
				}
			}
		})
	}
}

func TestOAuthAuthType(t *testing.T) {
	auth, _ := NewOAuthAuth("oauth_token_123", "")
	if got := auth.Type(); got != "oauth" {
		t.Errorf("OAuthAuth.Type() = %v, want 'oauth'", got)
	}
}

func TestOAuthAuthMask(t *testing.T) {
	auth, _ := NewOAuthAuth("oauth_token_123456789", "cloud-123")
	masked := auth.Mask()
	if strings.Contains(masked, "oauth_token_123456789") {
		t.Errorf("OAuthAuth.Mask() should not contain full token")
	}
	if !strings.Contains(masked, "****") {
		t.Errorf("OAuthAuth.Mask() should contain masked characters")
	}
	if !strings.Contains(masked, "cloud-123") {
		t.Errorf("OAuthAuth.Mask() should contain cloud ID")
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "empty token",
			token: "",
			want:  "<empty>",
		},
		{
			name:  "short token",
			token: "abc123",
			want:  "******",
		},
		{
			name:  "long token",
			token: "abcd1234567890xyz",
			want:  "abcd*********0xyz",
		},
		{
			name:  "exactly 8 chars",
			token: "12345678",
			want:  "********",
		},
		{
			name:  "9 chars",
			token: "123456789",
			want:  "1234*6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskToken(tt.token)
			if got != tt.want {
				t.Errorf("maskToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			apiKey:  "api-key-123",
			wantErr: false,
		},
		{
			name:    "missing API key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewAPIKeyAuth(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAPIKeyAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewAPIKeyAuth() returned nil auth without error")
			}
		})
	}
}

func TestAPIKeyAuthApply(t *testing.T) {
	auth, err := NewAPIKeyAuth("api-key-123")
	if err != nil {
		t.Fatalf("Failed to create APIKeyAuth: %v", err)
	}

	tests := []struct {
		name    string
		req     *http.Request
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &http.Request{Header: make(http.Header)},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.Apply(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIKeyAuth.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				authHeader := tt.req.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "GenieKey ") {
					t.Errorf("APIKeyAuth.Apply() Authorization header = %v, want prefix 'GenieKey '", authHeader)
				}
				if authHeader != "GenieKey api-key-123" {
					t.Errorf("APIKeyAuth.Apply() Authorization header = %v, want 'GenieKey api-key-123'", authHeader)
				}
			}
		})
	}
}

func TestAPIKeyAuthType(t *testing.T) {
	auth, _ := NewAPIKeyAuth("api-key-123")
	if got := auth.Type(); got != "apikey" {
		t.Errorf("APIKeyAuth.Type() = %v, want 'apikey'", got)
	}
}

func TestAPIKeyAuthMask(t *testing.T) {
	auth, _ := NewAPIKeyAuth("api-key-123456789")
	masked := auth.Mask()
	if strings.Contains(masked, "api-key-123456789") {
		t.Errorf("APIKeyAuth.Mask() should not contain full API key")
	}
	if !strings.Contains(masked, "****") {
		t.Errorf("APIKeyAuth.Mask() should contain masked characters")
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{
			name:   "empty API key",
			apiKey: "",
			want:   "<empty>",
		},
		{
			name:   "short API key",
			apiKey: "abc123",
			want:   "******",
		},
		{
			name:   "long API key",
			apiKey: "abcd1234567890xyz",
			want:   "abcd*********0xyz",
		},
		{
			name:   "exactly 8 chars",
			apiKey: "12345678",
			want:   "********",
		},
		{
			name:   "9 chars",
			apiKey: "123456789",
			want:   "1234*6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskAPIKey(tt.apiKey)
			if got != tt.want {
				t.Errorf("maskAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthError(t *testing.T) {
	baseErr := NewAuthError("base error", nil)
	wrappedErr := NewAuthError("wrapped error", baseErr)

	if baseErr.Error() != "base error" {
		t.Errorf("AuthError.Error() = %v, want 'base error'", baseErr.Error())
	}

	if !strings.Contains(wrappedErr.Error(), "wrapped error") {
		t.Errorf("AuthError.Error() should contain 'wrapped error'")
	}
}
