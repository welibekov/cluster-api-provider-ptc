package types

import "context"

// CertManager defines the methods for managing certificates.
type CertManager interface {
	// GetCert retrieves a certificate for the specified tenant ID.
	GetCert(ctx context.Context, tenantID string) ([]byte, error)
}
