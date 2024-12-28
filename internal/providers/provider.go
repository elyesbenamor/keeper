package providers

import (
	"context"
	"fmt"
	"time"
)

// Secret represents a secret value with metadata and versioning
type Secret struct {
	Value       string            `json:"value"`
	Metadata    map[string]string `json:"metadata"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
	Version     int              `json:"version"`
	PrevVersion *Secret          `json:"previous_version,omitempty"`
}

// Validate checks if the secret is valid
func (s *Secret) Validate() error {
	if s.Value == "" {
		return fmt.Errorf("secret value cannot be empty")
	}
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	if s.Created.IsZero() {
		s.Created = time.Now()
	}
	if s.Updated.IsZero() {
		s.Updated = time.Now()
	}
	if s.Version == 0 {
		s.Version = 1
	}
	return nil
}

// Provider is the interface that all secret providers must implement
type Provider interface {
	// Secret Management
	GetSecret(ctx context.Context, key string) (*Secret, error)
	SetSecret(ctx context.Context, key, value string, metadata map[string]string) error
	DeleteSecret(ctx context.Context, key string) error
	ListSecrets(ctx context.Context, prefix string) ([]string, error)
	
	// Version Management
	GetSecretVersion(ctx context.Context, key string, version int) (*Secret, error)
	ListSecretVersions(ctx context.Context, key string) ([]int, error)
	RollbackSecret(ctx context.Context, key string, version int) error
	
	// Search and Metadata
	SearchSecrets(ctx context.Context, query map[string]string) ([]*Secret, error)
	UpdateMetadata(ctx context.Context, key string, metadata map[string]string) error
	
	// Key Management
	RotateKey(ctx context.Context) error
	DeleteKey(ctx context.Context) error
	
	// Cleanup and Maintenance
	CleanupInvalidSecrets(ctx context.Context) error
	
	Close() error
}

// SearchOptions defines options for searching secrets
type SearchOptions struct {
	Metadata    map[string]string `json:"metadata"`
	SortBy      string           `json:"sort_by"`
	SortDesc    bool             `json:"sort_desc"`
	MaxResults  int              `json:"max_results"`
	IncludeKeys []string         `json:"include_keys"`
	ExcludeKeys []string         `json:"exclude_keys"`
}
