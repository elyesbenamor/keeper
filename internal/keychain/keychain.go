package keychain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/zalando/go-keyring"
)

const (
	servicePrefix = "keeper-"
	keyLength     = 32 // 256 bits
)

// Keychain is the interface for secure key storage
type Keychain interface {
	// Get retrieves a key by name
	Get(name string) ([]byte, error)

	// Set stores a key with the given name
	Set(name string, key []byte) error

	// Delete removes a key by name
	Delete(name string) error
}

// KeychainImpl manages encryption keys in the system keychain
type KeychainImpl struct {
	serviceName string
}

// New creates a new Keychain instance
func New() (*KeychainImpl, error) {
	return &KeychainImpl{
		serviceName: "keeper",
	}, nil
}

// GenerateKey generates a new encryption key and stores it in the keychain
func (k *KeychainImpl) GenerateKey(name string) error {
	// Generate random key
	key := make([]byte, keyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Store in keychain
	encoded := base64.StdEncoding.EncodeToString(key)
	err := keyring.Set(k.serviceName, name, encoded)
	if err != nil {
		return fmt.Errorf("failed to store key in keychain: %w", err)
	}

	return nil
}

// GetKey retrieves a key from the keychain
func (k *KeychainImpl) GetKey(name string) ([]byte, error) {
	encoded, err := keyring.Get(k.serviceName, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get key from keychain: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	return key, nil
}

// DeleteKey removes a key from the keychain
func (k *KeychainImpl) DeleteKey(name string) error {
	err := keyring.Delete(k.serviceName, name)
	if err != nil {
		return fmt.Errorf("failed to delete key from keychain: %w", err)
	}

	return nil
}

// RenameKey renames a key in the keychain
func (k *KeychainImpl) RenameKey(oldName, newName string) error {
	// Get the old key
	key, err := k.GetKey(oldName)
	if err != nil {
		return err
	}

	// Store with new name
	encoded := base64.StdEncoding.EncodeToString(key)
	if err := keyring.Set(k.serviceName, newName, encoded); err != nil {
		return fmt.Errorf("failed to store key with new name: %w", err)
	}

	// Delete old key
	if err := k.DeleteKey(oldName); err != nil {
		return err
	}

	return nil
}

// KeyExists checks if a key exists in the keychain
func (k *KeychainImpl) KeyExists(name string) (bool, error) {
	_, err := keyring.Get(k.serviceName, name)
	if err == keyring.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return true, nil
}

// Encrypt encrypts data using the specified key
func (k *KeychainImpl) Encrypt(keyName string, data []byte) ([]byte, error) {
	key, err := k.GetKey(keyName)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using the specified key
func (k *KeychainImpl) Decrypt(keyName string, data []byte) ([]byte, error) {
	key, err := k.GetKey(keyName)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// Get implements the Keychain interface
func (k *KeychainImpl) Get(name string) ([]byte, error) {
	return k.GetKey(name)
}

// Set implements the Keychain interface
func (k *KeychainImpl) Set(name string, key []byte) error {
	encoded := base64.StdEncoding.EncodeToString(key)
	return keyring.Set(k.serviceName, name, encoded)
}

// Delete implements the Keychain interface
func (k *KeychainImpl) Delete(name string) error {
	return k.DeleteKey(name)
}
