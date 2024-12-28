package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/keeper/internal/providers"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCPProvider implements the Provider interface using Google Cloud Secret Manager
type GCPProvider struct {
	client    *secretmanager.Client
	projectID string
}

// New creates a new GCPProvider
func New(ctx context.Context, projectID string, credentialsFile string) (*GCPProvider, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &GCPProvider{
		client:    client,
		projectID: projectID,
	}, nil
}

// GetSecret retrieves a secret from GCP Secret Manager
func (p *GCPProvider) GetSecret(ctx context.Context, key string) (*providers.Secret, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", p.projectID, key)
	result, err := p.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to access secret version: %w", err)
	}

	var secret providers.Secret
	if err := json.Unmarshal(result.Payload.Data, &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return &secret, nil
}

// SetSecret stores a secret in GCP Secret Manager
func (p *GCPProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	secret := &providers.Secret{
		Key:       key,
		Value:     value,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	// Create or update the secret
	secretPath := fmt.Sprintf("projects/%s/secrets/%s", p.projectID, key)
	_, err = p.client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", p.projectID),
		SecretId: key,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	if err != nil {
		// Ignore error if secret already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	// Add new version
	_, err = p.client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretPath,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
	}

	return nil
}

// DeleteSecret removes a secret from GCP Secret Manager
func (p *GCPProvider) DeleteSecret(ctx context.Context, key string) error {
	name := fmt.Sprintf("projects/%s/secrets/%s", p.projectID, key)
	err := p.client.DeleteSecret(ctx, &secretmanagerpb.DeleteSecretRequest{
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// ListSecrets lists all secrets in GCP Secret Manager with the given prefix
func (p *GCPProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	parent := fmt.Sprintf("projects/%s", p.projectID)
	it := p.client.ListSecrets(ctx, &secretmanagerpb.ListSecretsRequest{
		Parent: parent,
	})

	var keys []string
	for {
		secret, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		key := path.Base(secret.Name)
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// RotateSecret rotates a secret according to its rotation policy
func (p *GCPProvider) RotateSecret(ctx context.Context, key string) error {
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
func (p *GCPProvider) GetRotationPolicy(ctx context.Context, key string) (*providers.RotationPolicy, error) {
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
func (p *GCPProvider) SetRotationPolicy(ctx context.Context, key string, policy *providers.RotationPolicy) error {
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
