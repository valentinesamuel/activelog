package types

import (
	"context"
	"io"
	"time"
)

// PresignOperation represents the type of presigned URL operation
type PresignOperation string

const (
	PresignGet PresignOperation = "GET"
	PresignPut PresignOperation = "PUT"
)

// FileMetadata contains information about a stored file
type FileMetadata struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

// UploadInput contains parameters for uploading a file
type UploadInput struct {
	Key         string
	Body        io.Reader
	ContentType string
	Size        int64
	Metadata    map[string]string // Optional custom metadata
}

// UploadOutput contains the result of an upload operation
type UploadOutput struct {
	Key        string
	ETag       string
	URL        string // Full URL to access the object (if public)
	UploadedAt time.Time
}

// ListInput contains parameters for listing files
type ListInput struct {
	Prefix  string
	MaxKeys int
	Marker  string // For pagination
}

// ListOutput contains the result of a list operation
type ListOutput struct {
	Files       []FileMetadata
	NextMarker  string
	IsTruncated bool
}

// PresignedURLInput contains parameters for generating a presigned URL
type PresignedURLInput struct {
	Key       string
	ExpiresIn time.Duration
	Operation PresignOperation
}

// StorageProvider defines the interface for object storage operations
// All providers (S3, Supabase, Azure) must implement this interface
type StorageProvider interface {
	// Upload stores a file and returns metadata about the stored object
	Upload(ctx context.Context, input *UploadInput) (*UploadOutput, error)

	// Download retrieves a file by key
	// Caller is responsible for closing the returned ReadCloser
	Download(ctx context.Context, key string) (io.ReadCloser, *FileMetadata, error)

	// Delete removes a file by key
	// Returns nil if the file doesn't exist (idempotent)
	Delete(ctx context.Context, key string) error

	// DeleteMultiple removes multiple files efficiently
	// Returns a map of key -> error for any failures
	DeleteMultiple(ctx context.Context, keys []string) (map[string]error, error)

	// List returns files matching the given prefix
	List(ctx context.Context, input *ListInput) (*ListOutput, error)

	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)

	// GetPresignedURL generates a time-limited URL for direct access
	GetPresignedURL(ctx context.Context, input *PresignedURLInput) (string, error)

	// GetMetadata retrieves file metadata without downloading the file
	GetMetadata(ctx context.Context, key string) (*FileMetadata, error)
}
