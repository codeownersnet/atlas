package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// APIKeyAuth implements API Key authentication (Opsgenie GenieKey)
type APIKeyAuth struct {
	APIKey string
}

// NewAPIKeyAuth creates a new API key authentication provider
func NewAPIKeyAuth(apiKey string) (*APIKeyAuth, error) {
	if apiKey == "" {
		return nil, NewAuthError("API key is required", nil)
	}

	return &APIKeyAuth{
		APIKey: apiKey,
	}, nil
}

// Apply adds GenieKey authentication header to the request
func (a *APIKeyAuth) Apply(req *http.Request) error {
	if req == nil {
		return NewAuthError("request cannot be nil", nil)
	}

	// Opsgenie uses "GenieKey <api-key>" format
	req.Header.Set("Authorization", "GenieKey "+a.APIKey)

	return nil
}

// Type returns the authentication type
func (a *APIKeyAuth) Type() string {
	return "apikey"
}

// Mask returns a masked version of credentials for logging
func (a *APIKeyAuth) Mask() string {
	maskedKey := maskAPIKey(a.APIKey)
	return fmt.Sprintf("API key auth (key: %s)", maskedKey)
}

// maskAPIKey masks sensitive API key data for logging
// Shows first 4 and last 4 characters, masks the rest
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "<empty>"
	}
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}
