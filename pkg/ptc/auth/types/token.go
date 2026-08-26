package types

import (
	"golang.org/x/oauth2"
)

// ClientTokenManager is an interface for managing client tokens, including storing and loading tokens.
type ClientTokenManager interface {
	// StoreToken saves a token to a persistent storage.
	StoreToken([]byte) error

	// StoreAdminToken saves an admin token to a persistent storage.
	StoreAdminToken([]byte) error

	// LoadToken retrieves a stored token from persistent storage.
	LoadToken() (*oauth2.Token, error)

	// LoadAdminToken retrieves a stored admin token from persistent storage.
	LoadAdminToken() (*oauth2.Token, error)

	// DecodeToken extracts metadata from a given token.
	DecodeToken(*oauth2.Token) (*Principal, error)
}
