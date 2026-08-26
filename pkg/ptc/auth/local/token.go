package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

// Error definitions for TokenManager
var (
	// ErrNotLoggedIn is returned when a user attempts to perform an action that requires authentication
	// but the user is not logged in or does not have a valid session/token.
	ErrNotLoggedIn = errors.New("user is not logged in")
)

// TokenManager handles token-related operations.
type TokenManager struct {
	ctx context.Context
}

func NewFromConfig(ctx context.Context) *TokenManager {
	return &TokenManager{
		ctx: ctx,
	}
}

func (t *TokenManager) basePath() string {
	return filepath.Join(os.Getenv("HOME"), ".config/ptc")
}

func (t *TokenManager) tokenFilename() string {
	return "token.json"
}

func (t *TokenManager) adminTokenFilename() string {
	return "token_admin.json"
}
