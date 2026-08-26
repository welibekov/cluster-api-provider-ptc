package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PTCCredentials holds authenticated session details for PTC API.
type PTCCredentials struct {
	Username string
	Password string
	TenantID string
	ClientID string
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
	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	tenantID := string(secret.Data["tenantId"])
	clientID := string(secret.Data["clientId"])

	// Validate required fields
	if username == "" || password == "" || tenantID == "" || clientID == "" {
		return nil, fmt.Errorf("secret %s is missing required keys (username, password, tenantId, clientId)", secretKey)
	}

	return &PTCCredentials{
		Username: username,
		Password: password,
		TenantID: tenantID,
		ClientID: clientID,
	}, nil
}
