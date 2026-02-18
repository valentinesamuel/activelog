package types

import "errors"

// Storage error types
var (
	// ErrNotFound is returned when the requested object does not exist
	ErrNotFound = errors.New("storage: object not found")

	// ErrAccessDenied is returned when access to the object is denied
	ErrAccessDenied = errors.New("storage: access denied")

	// ErrInvalidKey is returned when the provided key is invalid
	ErrInvalidKey = errors.New("storage: invalid key")

	// ErrUploadFailed is returned when an upload operation fails
	ErrUploadFailed = errors.New("storage: upload failed")

	// ErrProviderNotConfigured is returned when the storage provider is not properly configured
	ErrProviderNotConfigured = errors.New("storage: provider not configured")

	// ErrUnsupportedProvider is returned when an unknown provider is requested
	ErrUnsupportedProvider = errors.New("storage: unsupported provider")
)
