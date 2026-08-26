package local

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth/types"
	"golang.org/x/oauth2"
)

func (t *TokenManager) DecodeToken(token *oauth2.Token) (*types.Principal, error) {
	var principal types.Principal

	parts := strings.Split(token.AccessToken, ".")

	if len(parts) < 2 {
		return nil, types.ErrInvalidToken
	}

	tokenBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("couldn't decode token: %v", err)
	}

	if err := json.Unmarshal(tokenBytes, &principal); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal token data: %v", err)
	}

	return &principal, nil
}
