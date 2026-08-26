package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"

	glAuth "github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth/auth"
	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/auth/types"

	"golang.org/x/oauth2"
)

type TokenOptionFunc func(*TokenOptions)

type TokenOptions struct {
	Refresh bool
	Store   bool
}

func WithoutRefresh() TokenOptionFunc {
	return func(to *TokenOptions) {
		to.Refresh = false
	}
}

func WithoutStore() TokenOptionFunc {
	return func(to *TokenOptions) {
		to.Store = false
	}
}

type TokenManager struct {
	scheme   string
	host     string
	basePath string
}

func NewTokenManager(scheme, host, basePath string) *TokenManager {
	return &TokenManager{
		scheme:   scheme,
		host:     host,
		basePath: basePath,
	}
}

// Store tokens.
func (t *TokenManager) StoreToken(tokenBytes []byte) error {
	return glAuth.GlobalClient.StoreToken(tokenBytes)
}

func (t *TokenManager) StoreTokenMust(token *oauth2.Token) {
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := t.StoreToken(tokenBytes); err != nil {
		log.Fatal(err)
	}
}

func (t *TokenManager) StoreAdminToken(tokenBytes []byte) error {
	return glAuth.GlobalClient.StoreAdminToken(tokenBytes)
}

func (t *TokenManager) StoreAdminTokenMust(token *oauth2.Token) {
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := t.StoreAdminToken(tokenBytes); err != nil {
		log.Fatal(err)
	}
}

// Load tokens.
func (t *TokenManager) LoadToken(opts ...TokenOptionFunc) (*oauth2.Token, error) {
	tknOpts := &TokenOptions{
		Refresh: true,
		Store:   true,
	}

	for _, opt := range opts {
		opt(tknOpts)
	}

	var (
		token *oauth2.Token
		err   error
	)

	token, err = glAuth.GlobalClient.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("couldn't get local token: %v", err)
	}

	if tknOpts.Refresh && !token.Valid() {
		token, err = t.RefreshToken(token)
		if err != nil {
			return nil, fmt.Errorf("couldn't refresh token: %v", err)
		}

		defer func() {
			if tknOpts.Store {
				t.StoreTokenMust(token)
			}
		}()
	}

	return token, nil
}

func (t *TokenManager) LoadTokenMust(opts ...TokenOptionFunc) *oauth2.Token {
	token, err := t.LoadToken(opts...)
	if err != nil {
		log.Fatal(err)
	}

	return token
}

func (t *TokenManager) LoadAdminToken() (*oauth2.Token, error) {
	return glAuth.GlobalClient.LoadAdminToken()
}

func (t *TokenManager) LoadAdminTokenMust() *oauth2.Token {
	token, err := t.LoadAdminToken()
	if err != nil {
		log.Fatal(err)
	}

	return token
}

func (t *TokenManager) DecodeToken(token *oauth2.Token) (*types.Principal, error) {
	return glAuth.GlobalClient.DecodeToken(token)
}

func (t *TokenManager) RefreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	//host := os.Getenv("PTC_CLIENT_HOST")
	//basePath := os.Getenv("PTC_CLIENT_BASE_PATH")
	//scheme := os.Getenv("PTC_CLIENT_SCHEME")

	// take params from token manager struct
	host := t.host
	basePath := t.basePath
	scheme := t.scheme

	tenantID := os.Getenv("PTC_AUTH_TENANT_ID")
	clientID := os.Getenv("PTC_AUTH_CLIENT_ID")
	clientSecret := os.Getenv("PTC_AUTH_CLIENT_SECRET")

	// Decode token and take some parameters from it.
	principal, err := t.DecodeToken(token)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", types.ErrTokenCannotBeParsed, err)
	}

	if principal.PtcClientID != "" {
		clientID = principal.PtcClientID
	}
	if principal.PtcTenantID != "" {
		tenantID = principal.PtcTenantID
	}

	cfg := client.DefaultTransportConfig().
		WithHost(host).
		WithBasePath(basePath).
		WithSchemes([]string{scheme})

	apiClient := client.NewHTTPClientWithConfig(nil, cfg)

	params := &operations.GetTokenParams{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: func(in string) *string { return &in }(clientSecret),
		GrantType:    "refresh_token",
		RefreshToken: &token.RefreshToken,
	}

	resp, err := apiClient.Operations.GetToken(params,
		operations.WithContentType("application/x-www-form-urlencoded"),
	)
	if err != nil {
		return nil, err
	}

	tokenJSON, err := json.Marshal(resp.Payload)
	if err != nil {
		return nil, fmt.Errorf("couldn't serialize token: %w", err)
	}

	var newToken oauth2.Token

	if err := json.Unmarshal(tokenJSON, &newToken); err != nil {
		return nil, err
	}

	return &newToken, nil
}
