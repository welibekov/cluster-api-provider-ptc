package types

import (
	"encoding/json"

	"golang.org/x/oauth2"
)

// Introspect holds information about token introspection.
type Introspect struct {
	Active bool     `json:"active"` // Indicates if the token is active
	Scope  []string `json:"scope"`  // The scope of the token
	Exp    int      `json:"exp"`    // The expiration time of the token
}

// IsValid checks if the token is valid.
func (i *Introspect) IsValid() bool {
	return i.Active
}

// GetBytes serializes the Introspect data to JSON.
func (i *Introspect) GetBytes() ([]byte, error) {
	return json.MarshalIndent(i, "", "  ")
}

// TokenRequestConfig holds configuration for token requests.
type TokenRequestConfig struct {
	Username string         // Username for authentication
	Password string         // Password for authentication
	Config   *oauth2.Config // OAuth2 configuration
	Token    *oauth2.Token  // OAuth2 token
}
