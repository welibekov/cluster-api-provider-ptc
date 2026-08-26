package local

import (
	"fmt"
	"os"
	"path/filepath"
)

func (t *TokenManager) StoreToken(tokenBytes []byte) error {
	return t.storeToken(t.tokenFilename(), tokenBytes)
}

func (t *TokenManager) StoreAdminToken(tokenBytes []byte) error {
	return t.storeToken(t.adminTokenFilename(), tokenBytes)
}

func (t *TokenManager) storeToken(tokenfile string, tokenBytes []byte) error {
	directory := t.basePath()

	// Create the client directory if it doesn't exist
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("couldn't create client directory %s: %v", directory, err)
	}

	// Write the token to a file
	if err := os.WriteFile(filepath.Join(directory, tokenfile), tokenBytes, 0o644); err != nil {
		return fmt.Errorf("couldn't store token: %v", err)
	}

	return nil
}
