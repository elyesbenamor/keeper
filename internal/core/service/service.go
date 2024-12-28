package service

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	vaultapi "github.com/hashicorp/vault/api"
	awsprovider "github.com/keeper/internal/providers/aws"
	"github.com/keeper/internal/providers/local"
	"github.com/keeper/internal/providers/vault"
	"github.com/keeper/internal/providers"
)

// Config represents the configuration for a secret provider
type Config struct {
	Type       string
	Parameters map[string]interface{}
}

// Service represents the secret management service
type Service struct {
	provider providers.Provider
}

// New creates a new Service with the given configuration
func New(cfg Config) (*Service, error) {
	var provider providers.Provider
	var err error

	switch cfg.Type {
	case "local":
		secretsDir, ok := cfg.Parameters["secretsDir"].(string)
		if !ok {
			return nil, fmt.Errorf("secretsDir parameter is required for local provider")
		}
		provider, err = local.New(secretsDir)

	case "vault":
		address, ok := cfg.Parameters["address"].(string)
		if !ok {
			return nil, fmt.Errorf("address parameter is required for vault provider")
		}
		token, ok := cfg.Parameters["token"].(string)
		if !ok {
			return nil, fmt.Errorf("token parameter is required for vault provider")
		}
		path, ok := cfg.Parameters["path"].(string)
		if !ok {
			return nil, fmt.Errorf("path parameter is required for vault provider")
		}

		vaultConfig := vaultapi.DefaultConfig()
		vaultConfig.Address = address
		if err := vaultConfig.ConfigureTLS(nil); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
		vaultConfig.ConfigureTLS(&vaultapi.TLSConfig{Insecure: true})

		client, err := vaultapi.NewClient(vaultConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create vault client: %w", err)
		}
		client.SetToken(token)

		provider, err = vault.New(client, path)
		if err != nil {
			return nil, fmt.Errorf("failed to create vault provider: %w", err)
		}

	case "aws":
		region, ok := cfg.Parameters["region"].(string)
		if !ok {
			return nil, fmt.Errorf("region parameter is required for aws provider")
		}

		awsConfig := aws.Config{
			Region: region,
		}
		provider, err = awsprovider.New(awsConfig)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &Service{
		provider: provider,
	}, nil
}

// Close closes the service and its provider
func (s *Service) Close() error {
	return s.provider.Close()
}

// GetSecret retrieves a secret
func (s *Service) GetSecret(ctx context.Context, key string) (*providers.Secret, error) {
	return s.provider.GetSecret(ctx, key)
}

// SetSecret stores a secret
func (s *Service) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	return s.provider.SetSecret(ctx, key, value, metadata)
}

// DeleteSecret removes a secret
func (s *Service) DeleteSecret(ctx context.Context, key string) error {
	return s.provider.DeleteSecret(ctx, key)
}

// ListSecrets lists all secrets with the given prefix
func (s *Service) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	return s.provider.ListSecrets(ctx, prefix)
}

// GetRotationPolicy retrieves the rotation policy for a secret
func (s *Service) GetRotationPolicy(ctx context.Context, key string) (*providers.RotationPolicy, error) {
	return s.provider.GetRotationPolicy(ctx, key)
}

// SetRotationPolicy sets the rotation policy for a secret
func (s *Service) SetRotationPolicy(ctx context.Context, key string, policy *providers.RotationPolicy) error {
	return s.provider.SetRotationPolicy(ctx, key, policy)
}

// RotateSecret rotates a secret according to its rotation policy
func (s *Service) RotateSecret(ctx context.Context, key string) error {
	return s.provider.RotateSecret(ctx, key)
}

// GetSecretMetadata gets metadata for a secret
func (s *Service) GetSecretMetadata(ctx context.Context, key string) (*providers.SecretMetadata, error) {
	return s.provider.GetSecretMetadata(ctx, key)
}
