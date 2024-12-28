package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/keeper/internal/providers"
)

// VaultProvider implements the Provider interface using HashiCorp Vault
type VaultProvider struct {
	client *api.Client
	path   string
}

// New creates a new VaultProvider
func New(client *api.Client, path string) (*VaultProvider, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client is required")
	}
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	return &VaultProvider{
		client: client,
		path:   path,
	}, nil
}

// GetSecret retrieves a secret from Vault
func (p *VaultProvider) GetSecret(ctx context.Context, key string) (*providers.Secret, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	secret, err := p.client.Logical().Read(path.Join(p.path, key))
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found: %s", key)
	}

	value, ok := secret.Data["value"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid secret value")
	}

	metadata := make(map[string]string)
	if metadataRaw, ok := secret.Data["metadata"].(map[string]interface{}); ok {
		for k, v := range metadataRaw {
			if str, ok := v.(string); ok {
				metadata[k] = str
			}
		}
	}

	return &providers.Secret{
		Value:    value,
		Metadata: metadata,
	}, nil
}

// SetSecret stores a secret in Vault
func (p *VaultProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}

	data := map[string]interface{}{
		"value":    value,
		"metadata": metadata,
	}

	_, err := p.client.Logical().Write(path.Join(p.path, key), data)
	if err != nil {
		return fmt.Errorf("failed to write secret: %w", err)
	}

	return nil
}

// DeleteSecret deletes a secret from Vault
func (p *VaultProvider) DeleteSecret(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}

	_, err := p.client.Logical().Delete(path.Join(p.path, key))
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// ListSecrets lists all secrets in Vault with the given prefix
func (p *VaultProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	secret, err := p.client.Logical().List(path.Join(p.path, prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var keys []string
	if secret == nil {
		return keys, nil
	}

	if keysList, ok := secret.Data["keys"].([]interface{}); ok {
		for _, key := range keysList {
			if str, ok := key.(string); ok {
				if strings.HasPrefix(str, prefix) {
					keys = append(keys, str)
				}
			}
		}
	}

	return keys, nil
}

// GetRotationPolicy retrieves the rotation policy for a secret
func (p *VaultProvider) GetRotationPolicy(ctx context.Context, key string) (*providers.RotationPolicy, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	secret, err := p.client.Logical().Read(path.Join(p.path, key, "rotation"))
	if err != nil {
		return nil, fmt.Errorf("failed to read rotation policy: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("rotation policy not found: %s", key)
	}

	policyData, ok := secret.Data["policy"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid rotation policy data")
	}

	var policy providers.RotationPolicy
	if err := json.Unmarshal([]byte(policyData), &policy); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rotation policy: %w", err)
	}

	return &policy, nil
}

// SetRotationPolicy sets the rotation policy for a secret
func (p *VaultProvider) SetRotationPolicy(ctx context.Context, key string, policy *providers.RotationPolicy) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}
	if policy == nil {
		return fmt.Errorf("policy is required")
	}

	policyData, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal rotation policy: %w", err)
	}

	data := map[string]interface{}{
		"policy": string(policyData),
	}

	_, err = p.client.Logical().Write(path.Join(p.path, key, "rotation"), data)
	if err != nil {
		return fmt.Errorf("failed to write rotation policy: %w", err)
	}

	return nil
}

// RotateSecret rotates a secret according to its rotation policy
func (p *VaultProvider) RotateSecret(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}

	policy, err := p.GetRotationPolicy(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get rotation policy: %w", err)
	}

	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Generate new value based on policy
	newValue := fmt.Sprintf("%d", providers.RandInt(policy.Length, policy.Length))

	return p.SetSecret(ctx, key, newValue, secret.Metadata)
}

// GetSecretMetadata gets metadata for a secret
func (p *VaultProvider) GetSecretMetadata(ctx context.Context, key string) (*providers.SecretMetadata, error) {
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return nil, err
	}

	// Get metadata from Vault
	metadata, err := p.client.Logical().ReadWithData(
		fmt.Sprintf("metadata/%s/%s", p.path, key),
		map[string][]string{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	var policy *providers.RotationPolicy
	if metadata != nil && metadata.Data != nil {
		if customMetadata, ok := metadata.Data["custom_metadata"].(map[string]interface{}); ok {
			if rotationData, ok := customMetadata["rotation_policy"].(map[string]interface{}); ok {
				interval, _ := time.ParseDuration(rotationData["interval"].(string))
				lastRotation, _ := time.Parse(time.RFC3339, rotationData["last_rotation"].(string))
				nextRotation, _ := time.Parse(time.RFC3339, rotationData["next_rotation"].(string))
				
				policy = &providers.RotationPolicy{
					Interval:     interval,
					LastRotation: lastRotation,
					NextRotation: nextRotation,
				}
			}
		}
	}

	return &providers.SecretMetadata{
		Created:         secret.Created,
		LastModified:    secret.Updated,
		UserMetadata:    secret.Metadata,
		RotationPolicy:  policy,
		PreviousVersion: "", // Vault handles versions internally
	}, nil
}

// Close implements the Provider interface
func (p *VaultProvider) Close() error {
	return nil
}
