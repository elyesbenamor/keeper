package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/keeper/internal/providers"
)

// AzureProvider implements the Provider interface using Azure Key Vault
type AzureProvider struct {
	client *azsecrets.Client
}

// New creates a new AzureProvider
func New(client *azsecrets.Client) (*AzureProvider, error) {
	return &AzureProvider{
		client: client,
	}, nil
}

// GetSecret retrieves a secret from Azure Key Vault
func (p *AzureProvider) GetSecret(ctx context.Context, key string) (*providers.Secret, error) {
	result, err := p.client.GetSecret(ctx, key, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	var secret providers.Secret
	if err := json.Unmarshal([]byte(*result.Value), &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return &secret, nil
}

// SetSecret stores a secret in Azure Key Vault
func (p *AzureProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	secret := &providers.Secret{
		Key:       key,
		Value:     value,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// If secret already exists, preserve creation time
	if existing, err := p.GetSecret(ctx, key); err == nil {
		secret.CreatedAt = existing.CreatedAt
	}

	data, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	params := azsecrets.SetSecretParameters{
		Value: &data,
	}

	_, err = p.client.SetSecret(ctx, key, params, nil)
	if err != nil {
		return fmt.Errorf("failed to set secret: %w", err)
	}

	return nil
}

// DeleteSecret removes a secret from Azure Key Vault
func (p *AzureProvider) DeleteSecret(ctx context.Context, key string) error {
	_, err := p.client.DeleteSecret(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// ListSecrets lists all secrets in Azure Key Vault with the given prefix
func (p *AzureProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	pager := p.client.NewListSecretsPager(nil)
	var keys []string

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, item := range page.Value {
			if item.ID == nil {
				continue
			}

			key := *item.ID
			if strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		}
	}

	return keys, nil
}

// RotateSecret rotates a secret according to its rotation policy
func (p *AzureProvider) RotateSecret(ctx context.Context, key string) error {
	// Get the existing secret
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get secret for rotation: %w", err)
	}

	// Get the rotation policy
	policy, err := p.GetRotationPolicy(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get rotation policy: %w", err)
	}

	// Check if it's time to rotate
	if time.Now().Before(policy.NextRotation) {
		return nil // Not time to rotate yet
	}

	// Generate new value
	var newValue string
	if policy.CustomGenerator != nil {
		newValue, err = policy.CustomGenerator()
		if err != nil {
			return fmt.Errorf("failed to generate new value: %w", err)
		}
	} else {
		newValue = fmt.Sprintf("rotated-%s-%d", secret.Value, time.Now().Unix())
	}

	// Update metadata
	if secret.Metadata == nil {
		secret.Metadata = make(map[string]string)
	}
	secret.Metadata["previous_value"] = secret.Value
	secret.Metadata["last_rotation"] = time.Now().Format(time.RFC3339)
	secret.Metadata["next_rotation"] = time.Now().Add(policy.Interval).Format(time.RFC3339)

	// Store the new value
	return p.SetSecret(ctx, key, newValue, secret.Metadata)
}

// GetRotationPolicy retrieves the rotation policy for a secret
func (p *AzureProvider) GetRotationPolicy(ctx context.Context, key string) (*providers.RotationPolicy, error) {
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return providers.DefaultRotationPolicy(), nil
	}

	if secret.Metadata == nil {
		return providers.DefaultRotationPolicy(), nil
	}

	// Try to parse rotation policy from metadata
	policyJSON, ok := secret.Metadata["rotation_policy"]
	if !ok {
		return providers.DefaultRotationPolicy(), nil
	}

	var policy providers.RotationPolicy
	if err := json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return providers.DefaultRotationPolicy(), nil
	}

	return &policy, nil
}

// SetRotationPolicy sets the rotation policy for a secret
func (p *AzureProvider) SetRotationPolicy(ctx context.Context, key string, policy *providers.RotationPolicy) error {
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	if secret.Metadata == nil {
		secret.Metadata = make(map[string]string)
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal rotation policy: %w", err)
	}

	secret.Metadata["rotation_policy"] = string(policyJSON)
	return p.SetSecret(ctx, key, secret.Value, secret.Metadata)
}
