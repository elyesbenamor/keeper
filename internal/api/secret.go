package api

// Secret represents a secret with its metadata
type Secret struct {
	// Key is the unique identifier for this secret
	Key string `json:"key"`

	// Value is the actual secret value
	Value string `json:"value"`

	// Metadata contains additional information about the secret
	Metadata map[string]string `json:"metadata,omitempty"`

	// Version is the version number of this secret
	Version int `json:"version"`

	// CreatedAt is the Unix timestamp when this secret was created
	CreatedAt int64 `json:"created_at"`

	// UpdatedAt is the Unix timestamp when this secret was last updated
	UpdatedAt int64 `json:"updated_at"`
}
