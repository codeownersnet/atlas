package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// BasicAuth implements Basic Authentication (username + API token)
type BasicAuth struct {
	Username string
	APIToken string
}

// NewBasicAuth creates a new basic authentication provider
func NewBasicAuth(username, apiToken string) (*BasicAuth, error) {
	if username == "" {
		return nil, NewAuthError("username is required for basic auth", nil)
	}
	if apiToken == "" {
		return nil, NewAuthError("API token is required for basic auth", nil)
	}

	return &BasicAuth{
		Username: username,
		APIToken: apiToken,
	}, nil
}

// Apply adds Basic Authentication header to the request
func (b *BasicAuth) Apply(req *http.Request) error {
	if req == nil {
		return NewAuthError("request cannot be nil", nil)
	}

	// Create basic auth header: "Basic base64(username:token)"
	credentials := fmt.Sprintf("%s:%s", b.Username, b.APIToken)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", "Basic "+encoded)

	return nil
}

// Type returns the authentication type
func (b *BasicAuth) Type() string {
	return "basic"
}

// Mask returns a masked version of credentials for logging
func (b *BasicAuth) Mask() string {
	maskedToken := maskToken(b.APIToken)
	return fmt.Sprintf("basic auth (user: %s, token: %s)", b.Username, maskedToken)
}

// maskToken masks sensitive token data for logging
// Shows first 4 and last 4 characters, masks the rest
func maskToken(token string) string {
	if token == "" {
		return "<empty>"
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
