package local

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

func (t *TokenManager) LoadToken() (*oauth2.Token, error) {
	return t.loadToken(t.tokenFilename())
}

func (t *TokenManager) LoadAdminToken() (*oauth2.Token, error) {
	return t.loadToken(t.adminTokenFilename())
}

func (t *TokenManager) loadToken(tokenfile string) (*oauth2.Token, error) {
	directory := t.basePath()

	tokenFileFullname := filepath.Join(directory, tokenfile)
	if checkTokenFile(tokenFileFullname) {
		return nil, ErrNotLoggedIn
	}

	// Read the token file into a byte slice.
	tokenBytes, err := os.ReadFile(tokenFileFullname)
	if err != nil {
		return nil, err // Return nil and the error if reading the file fails.
	}

	var token oauth2.Token

	// Unmarshal the byte slice into the token struct.
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal token: %v\n%s", err, string(tokenBytes)) // Return a formatted error if unmarshalling fails.
	}

	// Return the unmarshalled token.
	return &token, nil
}

func checkTokenFile(tokenFileFullname string) bool {
	// Check if file exists
	if _, err := os.Stat(tokenFileFullname); os.IsNotExist(err) {
		return true
	}
	return false
}
