package auth

import (
	"net/http"
)

// Provider is the interface for authentication providers
type Provider interface {
	// Apply adds authentication to an HTTP request
	Apply(req *http.Request) error

	// Type returns the authentication type
	Type() string

	// Mask returns a masked version of credentials for logging
	Mask() string
}

// Common error types
type AuthError struct {
	Message string
	Err     error
}

func (e *AuthError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError creates a new authentication error
func NewAuthError(message string, err error) error {
	return &AuthError{
		Message: message,
		Err:     err,
	}
}
