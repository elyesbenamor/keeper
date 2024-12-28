package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/keeper/internal/keychain"
	"github.com/keeper/internal/providers"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// LocalProvider implements the Provider interface using local filesystem storage
type LocalProvider struct {
	baseDir     string
	keychain    *keychain.Keychain
	mu          sync.RWMutex
	initialized bool
}

// New creates a new LocalProvider
func New(baseDir string) (*LocalProvider, error) {
	log.SetFormatter(&logrus.JSONFormatter{})
	
	kc, err := keychain.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize keychain: %w", err)
	}

	provider := &LocalProvider{
		baseDir:  baseDir,
		keychain: kc,
	}

	if err := provider.initialize(); err != nil {
		return nil, err
	}

	return provider, nil
}

func (p *LocalProvider) initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(p.baseDir, 0700); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create secrets directory if it doesn't exist
	secretsDir := filepath.Join(p.baseDir, "secrets")
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return fmt.Errorf("failed to create secrets directory: %w", err)
	}

	// Initialize master key if it doesn't exist
	if err := p.initializeMasterKey(); err != nil {
		return fmt.Errorf("failed to initialize master key: %w", err)
	}

	p.initialized = true
	return nil
}

func (p *LocalProvider) initializeMasterKey() error {
	exists, err := p.keychain.KeyExists("master")
	if err != nil {
		return fmt.Errorf("failed to check master key existence: %w", err)
	}

	if !exists {
		if err := p.keychain.GenerateKey("master"); err != nil {
			return fmt.Errorf("failed to generate master key: %w", err)
		}
		log.Info("Generated new master key")
	}

	return nil
}

// GetSecret retrieves a secret by key
func (p *LocalProvider) GetSecret(ctx context.Context, key string) (*providers.Secret, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	path := p.getSecretPath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("secret not found: %s", key)
		}
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}

	decrypted, err := p.keychain.Decrypt("master", data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	var secret providers.Secret
	if err := json.Unmarshal(decrypted, &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return &secret, nil
}

// SetSecret stores a secret with the given key and metadata
func (p *LocalProvider) SetSecret(ctx context.Context, key, value string, metadata map[string]string) error {
	// Get existing secret if it exists
	var prevVersion *providers.Secret
	existing, err := p.GetSecret(ctx, key)
	if err == nil {
		prevVersion = existing
	}

	// Create new secret
	secret := &providers.Secret{
		Value:    value,
		Metadata: metadata,
		Created:  time.Now(),
		Updated:  time.Now(),
		Version:  1,
	}

	if prevVersion != nil {
		secret.Version = prevVersion.Version + 1
		secret.Created = prevVersion.Created
		secret.PrevVersion = prevVersion
	}

	if err := secret.Validate(); err != nil {
		return fmt.Errorf("invalid secret: %w", err)
	}

	// Serialize and encrypt
	data, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	encrypted, err := p.keychain.Encrypt("master", data)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Lock only for file operations
	p.mu.Lock()
	defer p.mu.Unlock()

	// Save to file
	path := p.getSecretPath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create secret directory: %w", err)
	}

	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write secret file: %w", err)
	}

	log.WithFields(logrus.Fields{
		"key":     key,
		"version": secret.Version,
	}).Info("Secret set successfully")

	return nil
}

// DeleteSecret removes a secret by key
func (p *LocalProvider) DeleteSecret(ctx context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := p.getSecretPath(key)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("secret not found: %s", key)
		}
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	log.WithField("key", key).Info("Secret deleted successfully")
	return nil
}

// ListSecrets returns all secret keys with the given prefix
func (p *LocalProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var keys []string
	secretsDir := filepath.Join(p.baseDir, "secrets")
	err := filepath.Walk(secretsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			key := strings.TrimPrefix(path, secretsDir+"/")
			if strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return keys, nil
}

// GetSecretVersion retrieves a specific version of a secret
func (p *LocalProvider) GetSecretVersion(ctx context.Context, key string, version int) (*providers.Secret, error) {
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return nil, err
	}

	current := secret
	for current != nil {
		if current.Version == version {
			return current, nil
		}
		current = current.PrevVersion
	}

	return nil, fmt.Errorf("version %d not found for secret %s", version, key)
}

// ListSecretVersions returns all available versions for a secret
func (p *LocalProvider) ListSecretVersions(ctx context.Context, key string) ([]int, error) {
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return nil, err
	}

	var versions []int
	current := secret
	for current != nil {
		versions = append(versions, current.Version)
		current = current.PrevVersion
	}

	sort.Sort(sort.Reverse(sort.IntSlice(versions)))
	return versions, nil
}

// RollbackSecret reverts a secret to a specific version
func (p *LocalProvider) RollbackSecret(ctx context.Context, key string, version int) error {
	oldVersion, err := p.GetSecretVersion(ctx, key, version)
	if err != nil {
		return err
	}

	return p.SetSecret(ctx, key, oldVersion.Value, oldVersion.Metadata)
}

// SearchSecrets searches for secrets based on metadata
func (p *LocalProvider) SearchSecrets(ctx context.Context, query map[string]string) ([]*providers.Secret, error) {
	keys, err := p.ListSecrets(ctx, "")
	if err != nil {
		return nil, err
	}

	var results []*providers.Secret
	for _, key := range keys {
		secret, err := p.GetSecret(ctx, key)
		if err != nil {
			log.WithError(err).WithField("key", key).Warn("Failed to get secret during search")
			continue
		}

		matches := true
		for k, v := range query {
			if secret.Metadata[k] != v {
				matches = false
				break
			}
		}

		if matches {
			results = append(results, secret)
		}
	}

	return results, nil
}

// UpdateMetadata updates the metadata of a secret
func (p *LocalProvider) UpdateMetadata(ctx context.Context, key string, metadata map[string]string) error {
	secret, err := p.GetSecret(ctx, key)
	if err != nil {
		return err
	}

	return p.SetSecret(ctx, key, secret.Value, metadata)
}

// RotateKey rotates the master key and re-encrypts all secrets
func (p *LocalProvider) RotateKey(ctx context.Context) error {
	// List all secrets before locking
	keys, err := p.ListSecrets(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	// Get all secrets and their data before locking
	secretsData := make(map[string][]byte)
	for _, key := range keys {
		secret, err := p.GetSecret(ctx, key)
		if err != nil {
			log.WithError(err).WithField("key", key).Warn("Failed to get secret during key rotation")
			continue
		}

		data, err := json.Marshal(secret)
		if err != nil {
			log.WithError(err).WithField("key", key).Warn("Failed to marshal secret during key rotation")
			continue
		}

		secretsData[key] = data
	}

	// Now lock for the actual key rotation
	p.mu.Lock()
	defer p.mu.Unlock()

	// Generate new master key
	if err := p.keychain.GenerateKey("master_new"); err != nil {
		return fmt.Errorf("failed to generate new master key: %w", err)
	}

	// Re-encrypt all secrets with new key
	for key, data := range secretsData {
		encrypted, err := p.keychain.Encrypt("master_new", data)
		if err != nil {
			log.WithError(err).WithField("key", key).Warn("Failed to encrypt secret during key rotation")
			continue
		}

		path := p.getSecretPath(key)
		if err := os.WriteFile(path, encrypted, 0600); err != nil {
			log.WithError(err).WithField("key", key).Warn("Failed to write secret during key rotation")
			continue
		}
	}

	// Replace old key with new key
	if err := p.keychain.DeleteKey("master"); err != nil {
		return fmt.Errorf("failed to delete old master key: %w", err)
	}

	if err := p.keychain.RenameKey("master_new", "master"); err != nil {
		return fmt.Errorf("failed to rename new master key: %w", err)
	}

	log.Info("Master key rotated successfully")
	return nil
}

// DeleteKey deletes the master key
func (p *LocalProvider) DeleteKey(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.keychain.DeleteKey("master"); err != nil {
		return fmt.Errorf("failed to delete master key: %w", err)
	}

	log.Info("Master key deleted successfully")
	return nil
}

// CleanupInvalidSecrets removes any secrets that can't be decrypted or unmarshaled
func (p *LocalProvider) CleanupInvalidSecrets(ctx context.Context) error {
	keys, err := p.ListSecrets(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	var invalidKeys []string
	for _, key := range keys {
		_, err := p.GetSecret(ctx, key)
		if err != nil {
			invalidKeys = append(invalidKeys, key)
			log.WithError(err).WithField("key", key).Warn("Found invalid secret during cleanup")
		}
	}

	for _, key := range invalidKeys {
		if err := p.DeleteSecret(ctx, key); err != nil {
			log.WithError(err).WithField("key", key).Warn("Failed to delete invalid secret during cleanup")
		}
	}

	log.WithField("count", len(invalidKeys)).Info("Cleaned up invalid secrets")
	return nil
}

func (p *LocalProvider) getSecretPath(key string) string {
	return filepath.Join(p.baseDir, "secrets", key)
}

// Close implements the Provider interface
func (p *LocalProvider) Close() error {
	return nil
}
