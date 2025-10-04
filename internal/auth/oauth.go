package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// OAuthAuth implements OAuth 2.0 Bearer token authentication
// This is a BYO (Bring Your Own) token implementation - users must provide their own access tokens
type OAuthAuth struct {
	AccessToken string
	CloudID     string
}

// NewOAuthAuth creates a new OAuth authentication provider with a pre-existing access token
// Users must obtain the access token through their own OAuth flow or from Atlassian admin
func NewOAuthAuth(accessToken, cloudID string) (*OAuthAuth, error) {
	if accessToken == "" {
		return nil, NewAuthError("access token is required for OAuth", nil)
	}

	return &OAuthAuth{
		AccessToken: accessToken,
		CloudID:     cloudID,
	}, nil
}

// Apply adds OAuth Bearer token header to the request
func (o *OAuthAuth) Apply(req *http.Request) error {
	if req == nil {
		return NewAuthError("request cannot be nil", nil)
	}

	// OAuth uses Bearer token authentication
	req.Header.Set("Authorization", "Bearer "+o.AccessToken)

	// Add cloud ID header if provided (for multi-cloud support)
	if o.CloudID != "" {
		req.Header.Set("X-Atlassian-Cloud-Id", o.CloudID)
	}

	return nil
}

// Type returns the authentication type
func (o *OAuthAuth) Type() string {
	return "oauth"
}

// Mask returns a masked version of credentials for logging
func (o *OAuthAuth) Mask() string {
	maskedToken := maskOAuthToken(o.AccessToken)
	cloudInfo := ""
	if o.CloudID != "" {
		cloudInfo = fmt.Sprintf(", cloud_id: %s", o.CloudID)
	}
	return fmt.Sprintf("OAuth auth (token: %s%s)", maskedToken, cloudInfo)
}

// maskOAuthToken masks sensitive token data for logging
// Shows first 4 and last 4 characters, masks the rest
func maskOAuthToken(token string) string {
	if token == "" {
		return "<empty>"
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
