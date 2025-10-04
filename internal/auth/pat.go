package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// PATAuth implements Personal Access Token authentication
type PATAuth struct {
	Token string
}

// NewPATAuth creates a new PAT authentication provider
func NewPATAuth(token string) (*PATAuth, error) {
	if token == "" {
		return nil, NewAuthError("personal access token is required", nil)
	}

	return &PATAuth{
		Token: token,
	}, nil
}

// Apply adds Bearer token header to the request
func (p *PATAuth) Apply(req *http.Request) error {
	if req == nil {
		return NewAuthError("request cannot be nil", nil)
	}

	// PAT uses Bearer token authentication
	req.Header.Set("Authorization", "Bearer "+p.Token)

	return nil
}

// Type returns the authentication type
func (p *PATAuth) Type() string {
	return "pat"
}

// Mask returns a masked version of credentials for logging
func (p *PATAuth) Mask() string {
	maskedToken := maskPATToken(p.Token)
	return fmt.Sprintf("PAT auth (token: %s)", maskedToken)
}

// maskPATToken masks sensitive token data for logging
// Shows first 4 and last 4 characters, masks the rest
func maskPATToken(token string) string {
	if token == "" {
		return "<empty>"
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
