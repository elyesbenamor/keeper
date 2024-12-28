package local

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/keeper/internal/keychain"
	"github.com/keeper/internal/providers"
)

// LocalProvider implements the Provider interface using local filesystem storage
type LocalProvider struct {
	baseDir   string
	backupDir string
	keychain  keychain.Keychain
	mu        sync.RWMutex
}

// New creates a new LocalProvider
func New(baseDir string, kc keychain.Keychain) (*LocalProvider, error) {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalProvider{
		baseDir:  baseDir,
		keychain: kc,
	}, nil
}

// Initialize initializes the provider
func (p *LocalProvider) Initialize(ctx context.Context) error {
	// Create necessary directories
	dirs := []string{
		filepath.Join(p.baseDir, "secrets"),
		filepath.Join(p.baseDir, "schemas"),
		filepath.Join(p.baseDir, "backups"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Close closes the provider
func (p *LocalProvider) Close() error {
	return nil
}

// GetSecret retrieves a secret by name
func (p *LocalProvider) GetSecret(ctx context.Context, name string) (*providers.Secret, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	path := filepath.Join(p.baseDir, "secrets", name+".json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, providers.ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}

	var secret providers.Secret
	if err := json.Unmarshal(data, &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return &secret, nil
}

// SetSecret stores a secret
func (p *LocalProvider) SetSecret(ctx context.Context, secret *providers.Secret) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Load schema if specified
	if secret.Schema != "" {
		schemaPath := filepath.Join(p.baseDir, "schemas", secret.Schema+".json")
		schema, err := providers.LoadSchema(schemaPath)
		if err != nil {
			return fmt.Errorf("failed to load schema: %w", err)
		}

		// Validate secret against schema
		if err := providers.ValidateSecret(secret, schema); err != nil {
			return fmt.Errorf("secret validation failed: %w", err)
		}
	}

	// Validate secret
	if err := secret.Validate(); err != nil {
		return fmt.Errorf("invalid secret: %w", err)
	}

	// Update timestamps
	now := time.Now()
	if secret.CreatedAt.IsZero() {
		secret.CreatedAt = now
	}
	secret.UpdatedAt = now

	// Marshal secret
	data, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	// Write to file
	path := filepath.Join(p.baseDir, "secrets", secret.Name+".json")
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write secret file: %w", err)
	}

	return nil
}

// DeleteSecret deletes a secret by name
func (p *LocalProvider) DeleteSecret(ctx context.Context, name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := filepath.Join(p.baseDir, "secrets", name+".json")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return providers.ErrSecretNotFound
		}
		return fmt.Errorf("failed to delete secret file: %w", err)
	}

	return nil
}

// ListSecrets lists all secrets
func (p *LocalProvider) ListSecrets(ctx context.Context) ([]*providers.Secret, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	dir := filepath.Join(p.baseDir, "secrets")
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets directory: %w", err)
	}

	var secrets []*providers.Secret
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read secret file %s: %w", file.Name(), err)
		}

		var secret providers.Secret
		if err := json.Unmarshal(data, &secret); err != nil {
			return nil, fmt.Errorf("failed to unmarshal secret %s: %w", file.Name(), err)
		}

		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

// SearchSecrets searches for secrets based on criteria
func (p *LocalProvider) SearchSecrets(ctx context.Context, opts providers.SearchOptions) ([]*providers.Secret, error) {
	secrets, err := p.ListSecrets(ctx)
	if err != nil {
		return nil, err
	}

	var results []*providers.Secret
	for _, secret := range secrets {
		if !matchesSearch(secret, opts) {
			continue
		}
		results = append(results, secret)
	}

	return results, nil
}

// matchesSearch checks if a secret matches the search criteria
func matchesSearch(secret *providers.Secret, opts providers.SearchOptions) bool {
	// Check schema
	if opts.Schema != "" && secret.Schema != opts.Schema {
		return false
	}

	// Check tags
	if len(opts.Tags) > 0 {
		found := false
		for _, tag := range opts.Tags {
			for _, secretTag := range secret.Tags {
				if tag == secretTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Check created after
	if !opts.CreatedAfter.IsZero() && secret.CreatedAt.Before(opts.CreatedAfter) {
		return false
	}

	return true
}

// SetBackupDir sets the backup directory
func (p *LocalProvider) SetBackupDir(dir string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.backupDir = dir
	return nil
}

// Backup creates a backup of all secrets
func (p *LocalProvider) Backup(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.backupDir == "" {
		return fmt.Errorf("backup directory not set")
	}

	// List all secrets
	secrets, err := p.ListSecrets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(p.backupDir, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup each secret
	for _, secret := range secrets {
		data, err := json.MarshalIndent(secret, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal secret %s: %w", secret.Name, err)
		}

		path := filepath.Join(p.backupDir, secret.Name+".json")
		if err := ioutil.WriteFile(path, data, 0600); err != nil {
			return fmt.Errorf("failed to write backup file: %w", err)
		}
	}

	return nil
}

// Restore restores secrets from a backup
func (p *LocalProvider) Restore(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.backupDir == "" {
		return fmt.Errorf("backup directory not set")
	}

	// Read backup directory
	files, err := ioutil.ReadDir(p.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Restore each secret
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(p.backupDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read backup file %s: %w", file.Name(), err)
		}

		var secret providers.Secret
		if err := json.Unmarshal(data, &secret); err != nil {
			return fmt.Errorf("failed to unmarshal backup file %s: %w", file.Name(), err)
		}

		if err := p.SetSecret(ctx, &secret); err != nil {
			return fmt.Errorf("failed to restore secret %s: %w", secret.Name, err)
		}
	}

	return nil
}
