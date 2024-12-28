package api

import (
	"context"
)

// Secret represents a secret with its metadata
type Secret struct {
	Key       string
	Value     string
	Version   int
	Metadata  map[string]string
	CreatedAt int64
	UpdatedAt int64
}

// Provider defines the interface that all secret providers must implement
type Provider interface {
	// Initialize sets up the provider with configuration
	Initialize(config map[string]interface{}) error

	// Get retrieves a secret by its key
	Get(ctx context.Context, key string) (*Secret, error)

	// Set stores a secret
	Set(ctx context.Context, key string, value string, metadata map[string]string) error

	// Delete removes a secret
	Delete(ctx context.Context, key string) error

	// List returns all secret keys under a path
	List(ctx context.Context, path string) ([]string, error)

	// GetVersion retrieves a specific version of a secret
	GetVersion(ctx context.Context, key string, version int) (*Secret, error)

	// Close cleans up any resources used by the provider
	Close() error
}
