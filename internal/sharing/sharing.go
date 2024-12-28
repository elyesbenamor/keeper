package sharing

import (
	"context"
	"fmt"
	"time"

	"github.com/keeper/internal/providers"
)

// ShareRequest represents a request to share a secret between providers
type ShareRequest struct {
	Key            string
	SourceProvider providers.Provider
	TargetProvider providers.Provider
	ExpiresAt      *time.Time
}

// ShareSecret shares a secret from the source provider to the target provider
func ShareSecret(ctx context.Context, req *ShareRequest) error {
	// Get the secret from source provider
	secret, err := req.SourceProvider.GetSecret(ctx, req.Key)
	if err != nil {
		return fmt.Errorf("failed to get secret from source: %w", err)
	}

	// Check if the secret has expired
	if secret.Metadata != nil {
		if expiresStr, ok := secret.Metadata["expires_at"]; ok {
			expires, err := time.Parse(time.RFC3339, expiresStr)
			if err == nil && time.Now().After(expires) {
				return fmt.Errorf("secret has expired")
			}
		}
	}

	// Create metadata for shared secret
	metadata := make(map[string]string)
	if secret.Metadata != nil {
		for k, v := range secret.Metadata {
			metadata[k] = v
		}
	}
	metadata["shared_from"] = req.Key
	metadata["shared_at"] = time.Now().Format(time.RFC3339)

	if req.ExpiresAt != nil {
		metadata["expires_at"] = req.ExpiresAt.Format(time.RFC3339)
	}

	// Share the secret to target provider
	err = req.TargetProvider.SetSecret(ctx, req.Key, secret.Value, metadata)
	if err != nil {
		return fmt.Errorf("failed to set secret in target: %w", err)
	}

	return nil
}

// SyncSharedSecret synchronizes a previously shared secret
func SyncSharedSecret(ctx context.Context, req *ShareRequest) error {
	// Get the secret from target to verify it was shared
	targetSecret, err := req.TargetProvider.GetSecret(ctx, req.Key)
	if err != nil {
		return fmt.Errorf("failed to get secret from target: %w", err)
	}

	// Verify this is a shared secret
	if targetSecret.Metadata == nil || targetSecret.Metadata["shared_from"] == "" {
		return fmt.Errorf("secret was not previously shared")
	}

	// Get the source secret
	sourceSecret, err := req.SourceProvider.GetSecret(ctx, req.Key)
	if err != nil {
		return fmt.Errorf("failed to get secret from source: %w", err)
	}

	// Update metadata
	metadata := make(map[string]string)
	for k, v := range targetSecret.Metadata {
		metadata[k] = v
	}
	metadata["synced_at"] = time.Now().Format(time.RFC3339)

	// Sync the secret
	err = req.TargetProvider.SetSecret(ctx, req.Key, sourceSecret.Value, metadata)
	if err != nil {
		return fmt.Errorf("failed to sync secret in target: %w", err)
	}

	return nil
}

// RevokeSharing revokes a previously shared secret
func RevokeSharing(ctx context.Context, req *ShareRequest) error {
	// Get the secret from target to verify it was shared
	targetSecret, err := req.TargetProvider.GetSecret(ctx, req.Key)
	if err != nil {
		return fmt.Errorf("failed to get secret from target: %w", err)
	}

	// Verify this is a shared secret
	if targetSecret.Metadata == nil || targetSecret.Metadata["shared_from"] == "" {
		return fmt.Errorf("secret was not previously shared")
	}

	// Delete the secret from target
	err = req.TargetProvider.DeleteSecret(ctx, req.Key)
	if err != nil {
		return fmt.Errorf("failed to delete secret from target: %w", err)
	}

	return nil
}
