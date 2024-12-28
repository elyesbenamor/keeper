package providers

import (
	"context"
	"errors"
	"time"
)

// ErrSecretNotFound is returned when a secret is not found
var ErrSecretNotFound = errors.New("secret not found")

// Secret represents a secret with metadata
type Secret struct {
	Name      string            `json:"name"`
	Value     string            `json:"value"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Schema    string            `json:"schema,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// SearchOptions represents options for searching secrets
type SearchOptions struct {
	Schema        string    `json:"schema,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	CreatedAfter time.Time `json:"created_after,omitempty"`
}

// Provider defines the interface for secret management
type Provider interface {
	// Initialize initializes the provider
	Initialize(ctx context.Context) error

	// Close closes the provider
	Close() error

	// GetSecret retrieves a secret by name
	GetSecret(ctx context.Context, name string) (*Secret, error)

	// SetSecret stores a secret
	SetSecret(ctx context.Context, secret *Secret) error

	// DeleteSecret deletes a secret by name
	DeleteSecret(ctx context.Context, name string) error

	// ListSecrets lists all secrets
	ListSecrets(ctx context.Context) ([]*Secret, error)

	// SearchSecrets searches for secrets based on criteria
	SearchSecrets(ctx context.Context, opts SearchOptions) ([]*Secret, error)

	// SetBackupDir sets the backup directory for the provider
	SetBackupDir(dir string) error

	// Backup creates a backup of all secrets
	Backup(ctx context.Context) error

	// Restore restores secrets from a backup
	Restore(ctx context.Context) error
}

// Validate checks if the secret is valid
func (s *Secret) Validate() error {
	if s.Name == "" {
		return errors.New("secret name cannot be empty")
	}
	if s.Value == "" {
		return errors.New("secret value cannot be empty")
	}
	return nil
}

// NewSecret creates a new secret with the given name and value
func NewSecret(name, value string) *Secret {
	now := time.Now()
	return &Secret{
		Name:      name,
		Value:     value,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}
