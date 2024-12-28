package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/keeper/internal/keychain"
)

// Cipher handles encryption and decryption of data
type Cipher struct {
	gcm cipher.AEAD
}

// NewCipher creates a new cipher using the master key from keychain
func NewCipher(appDir string) (*Cipher, error) {
	// Get key from keychain
	km := keychain.New(appDir)
	key, err := km.GetMasterKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get master key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Cipher{gcm: gcm}, nil
}

// Encrypt encrypts data using AES-GCM
func (c *Cipher) Encrypt(data []byte) (string, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	encrypted := c.gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts data using AES-GCM
func (c *Cipher) Decrypt(encryptedStr string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := c.gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
