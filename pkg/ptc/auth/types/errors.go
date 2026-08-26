// Package types defines predefined errors for tenant and user operations.
package types

import "fmt"

// Predefined errors for tenant and user operations.
var (
	// ErrNoTenantFound indicates that no tenant could be found.
	ErrNoTenantFound = fmt.Errorf("no tenant found")

	// ErrNoUserFound indicates that no user could be found.
	ErrNoUserFound = fmt.Errorf("no user found")

	// ErrTenantAlreadyExists indicates that the tenant already exists.
	ErrTenantAlreadyExists = fmt.Errorf("tenant already exists")

	// ErrUserAlreadyExists indicates that the user already exists.
	ErrUserAlreadyExists = fmt.Errorf("user already exists")

	// ErrNoClientIDFound indicates that no client_id could be found.
	ErrNoClientIDFound = fmt.Errorf("no client_id found")

	// ErrInvalidCredentials indicates that the provided credentials are invalid.
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")

	// ErrFailedToHashPassword indicates that there was an error hashing the password.
	ErrFailedToHashPassword = fmt.Errorf("failed to hash password")

	// ErrAuthenticationFailed indicates that authentication has failed.
	ErrAuthenticationFailed = fmt.Errorf("authentication failed")

	// ErrInvalidToken indicates that the provided token is not valid.
	ErrInvalidToken = fmt.Errorf("token is not valid")

	// ErrTokenCannotBeVerified indicates that the token cannot be verified.
	ErrTokenCannotBeVerified = fmt.Errorf("token cannot be verified")

	// ErrTokenCannotBeParsed indicates that the token cannot be parsed.
	ErrTokenCannotBeParsed = fmt.Errorf("token cannot be parsed")

	// ErrFailedToCreateTenant indicates that there was an error creating the tenant.
	ErrFailedToCreateTenant = fmt.Errorf("failed to create tenant")

	// ErrFailedToCreateUser indicates that there was an error creating the user.
	ErrFailedToCreateUser = fmt.Errorf("failed to create user")

	// ErrUserLocked indicates that the user account is locked.
	ErrUserLocked = fmt.Errorf("user is locked")

	// ErrTenantLocked indicates that the tenant is locked.
	ErrTenantLocked = fmt.Errorf("tenant is locked")
)
