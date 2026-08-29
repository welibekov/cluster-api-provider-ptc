package controller

import (
	"context"
	"fmt"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PTCCredentials holds authenticated session details for PTC API.
type PTCCredentials struct {
	OriginalHostname string
	Scheme           string
	Hostname         string
	Username         string
	Password         string
	TenantID         string
	ClientID         string
}

// FetchCredentials reads and validates the PTC secret from Kubernetes API.
func FetchCredentials(ctx context.Context, c client.Client, ref *corev1.SecretReference, defaultNamespace string) (*PTCCredentials, error) {
	if ref == nil {
		return nil, fmt.Errorf("identityRef is nil")
	}

	namespace := ref.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}

	secret := &corev1.Secret{}
	secretKey := types.NamespacedName{
		Namespace: namespace,
		Name:      ref.Name,
	}

	if err := c.Get(ctx, secretKey, secret); err != nil {
		return nil, fmt.Errorf("failed to fetch secret %s: %w", secretKey, err)
	}

	// Extract values from secret.Data ([]byte map)
	hostname := string(secret.Data["ptcHostname"])
	username := string(secret.Data["ptcUsername"])
	password := string(secret.Data["ptcPassword"])
	tenantID := string(secret.Data["ptcTenantId"])
	clientID := string(secret.Data["ptcClientId"])

	// Validate required fields
	if username == "" || password == "" || tenantID == "" || hostname == "" || clientID == "" {
		return nil, fmt.Errorf("secret %s is missing required keys (ptcUsername, ptcPassword, ptcTenantId, ptcHostname, ptcClientID)", secretKey)
	}

	parsedURL, err := parseURL(hostname)
	if err != nil {
		return nil, fmt.Errorf("could not parse hostname %s: %w", hostname, err)
	}

	return &PTCCredentials{
		Scheme:           parsedURL.Scheme,
		Hostname:         parsedURL.Host,
		Username:         username,
		Password:         password,
		TenantID:         tenantID,
		ClientID:         clientID,
		OriginalHostname: hostname,
	}, nil
}

// parseURL takes a URL string and returns a pointer to a url.URL.
func parseURL(urlStr string) (*url.URL, error) {
	// Parse the URL using net/url
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Verify that the scheme and host are not empty
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("parsed URL is incomplete: scheme or host is empty")
	}

	return parsedURL, nil
}
