package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/keeper/internal/providers"
	"github.com/pkg/errors"
)

// AWSProvider implements the Provider interface using AWS Secrets Manager
type AWSProvider struct {
	client *secretsmanager.Client
}

// New creates a new AWSProvider with the given configuration
func New(cfg aws.Config) (*AWSProvider, error) {
	client := secretsmanager.NewFromConfig(cfg)
	return &AWSProvider{
		client: client,
	}, nil
}

// GetSecret retrieves a secret from AWS Secrets Manager
func (p *AWSProvider) GetSecret(ctx context.Context, key string) (*providers.Secret, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}

	result, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	var secret providers.Secret
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return &secret, nil
}

// SetSecret stores a secret in AWS Secrets Manager
func (p *AWSProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	now := time.Now()
	secret := providers.Secret{
		Value:    value,
		Metadata: metadata,
		Created:  now,
		Updated:  now,
	}

	data, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(key),
		SecretString: aws.String(string(data)),
	}

	_, err = p.client.CreateSecret(ctx, input)
	if err != nil {
		var alreadyExists *types.ResourceExistsException
		if errors.As(err, &alreadyExists) {
			// Secret already exists, update it
			updateInput := &secretsmanager.UpdateSecretInput{
				SecretId:     aws.String(key),
				SecretString: aws.String(string(data)),
			}
			_, err = p.client.UpdateSecret(ctx, updateInput)
			if err != nil {
				return fmt.Errorf("failed to update secret: %w", err)
			}
		} else {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	return nil
}

// DeleteSecret deletes a secret from AWS Secrets Manager
func (p *AWSProvider) DeleteSecret(ctx context.Context, key string) error {
	input := &secretsmanager.DeleteSecretInput{
		SecretId: aws.String(key),
	}

	_, err := p.client.DeleteSecret(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// ListSecrets lists all secrets in AWS Secrets Manager
func (p *AWSProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	input := &secretsmanager.ListSecretsInput{}
	result, err := p.client.ListSecrets(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var keys []string
	for _, secret := range result.SecretList {
		if prefix == "" || aws.ToString(secret.Name) == prefix {
			keys = append(keys, aws.ToString(secret.Name))
		}
	}

	return keys, nil
}

// GetRotationPolicy retrieves the rotation policy for a secret
func (p *AWSProvider) GetRotationPolicy(ctx context.Context, key string) (*providers.RotationPolicy, error) {
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	if secret.Metadata == nil {
		return nil, nil
	}

	policyJSON, ok := secret.Metadata["rotation_policy"]
	if !ok {
		return nil, nil
	}

	var policy providers.RotationPolicy
	if err := json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rotation policy: %w", err)
	}

	return &policy, nil
}

// SetRotationPolicy sets the rotation policy for a secret
func (p *AWSProvider) SetRotationPolicy(ctx context.Context, key string, policy *providers.RotationPolicy) error {
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

// RotateSecret rotates a secret according to its rotation policy
func (p *AWSProvider) RotateSecret(ctx context.Context, key string) error {
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
func (p *AWSProvider) GetSecretMetadata(ctx context.Context, key string) (*providers.SecretMetadata, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}

	output, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	var secret providers.Secret
	if err := json.Unmarshal([]byte(*output.SecretString), &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	// Get rotation policy if exists
	var policy *providers.RotationPolicy
	describeInput := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(key),
	}
	describeOutput, err := p.client.DescribeSecret(ctx, describeInput)
	if err == nil && describeOutput.RotationEnabled != nil && *describeOutput.RotationEnabled {
		policy = &providers.RotationPolicy{
			Interval:     time.Duration(*describeOutput.RotationRules.AutomaticallyAfterDays) * 24 * time.Hour,
			LastRotation: *describeOutput.LastRotatedDate,
			NextRotation: *describeOutput.NextRotationDate,
		}
	}

	return &providers.SecretMetadata{
		Created:         secret.Created,
		LastModified:    secret.Updated,
		UserMetadata:    secret.Metadata,
		RotationPolicy:  policy,
		PreviousVersion: "", // AWS Secrets Manager handles versions internally
	}, nil
}

// Close implements the Provider interface
func (p *AWSProvider) Close() error {
	return nil
}
