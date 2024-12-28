package local

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keeper/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProvider_BasicSecretOperations(t *testing.T) {
	log.Printf("Starting TestLocalProvider_BasicSecretOperations")
	
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "keeper-local-test")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Created temp directory: %s", tempDir)
	defer func() {
		log.Printf("Cleaning up temp directory: %s", tempDir)
		os.RemoveAll(tempDir)
	}()

	// Create provider
	secretsDir := filepath.Join(tempDir, "secrets")
	log.Printf("Creating provider with secrets directory: %s", secretsDir)
	provider, err := New(secretsDir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("Set and Get Secret", func(t *testing.T) {
		log.Printf("Running Set and Get Secret test")
		key := "app/db/password"
		value := "mysecretpassword"
		metadata := map[string]string{
			"environment": "production",
			"owner":      "dbadmin",
		}

		// Set a secret
		log.Printf("Setting secret with key: %s", key)
		err := provider.SetSecret(ctx, key, value, metadata)
		if err != nil {
			t.Fatalf("Failed to set secret: %v", err)
		}

		// Get the secret back
		log.Printf("Getting secret with key: %s", key)
		secret, err := provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get secret: %v", err)
		}

		// Verify secret
		assert.Equal(t, value, secret.Value)
		assert.Equal(t, metadata["environment"], secret.Metadata["environment"])
		assert.Equal(t, metadata["owner"], secret.Metadata["owner"])
		log.Printf("Successfully verified secret and metadata")

		// Cleanup
		provider.DeleteSecret(ctx, key)
	})

	t.Run("Update Secret", func(t *testing.T) {
		log.Printf("Running Update Secret test")
		key := "app/api/key"
		
		// Set initial secret
		log.Printf("Setting initial secret with key: %s", key)
		err := provider.SetSecret(ctx, key, "initial-value", nil)
		if err != nil {
			t.Fatalf("Failed to set initial secret: %v", err)
		}

		// Get initial timestamp
		log.Printf("Getting initial secret")
		initial, err := provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get initial secret: %v", err)
		}

		// Wait a bit to ensure timestamps are different
		time.Sleep(10 * time.Millisecond)

		// Update secret
		log.Printf("Updating secret")
		err = provider.SetSecret(ctx, key, "updated-value", map[string]string{"updated": "true"})
		if err != nil {
			t.Fatalf("Failed to update secret: %v", err)
		}

		// Get updated secret
		log.Printf("Getting updated secret")
		updated, err := provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get updated secret: %v", err)
		}

		// Verify changes
		assert.Equal(t, "updated-value", updated.Value)
		assert.Equal(t, "true", updated.Metadata["updated"])
		assert.Equal(t, initial.CreatedAt, updated.CreatedAt)
		assert.True(t, updated.UpdatedAt.After(initial.UpdatedAt))
		log.Printf("Successfully verified updated secret")

		// Cleanup
		provider.DeleteSecret(ctx, key)
	})

	t.Run("Delete Secret", func(t *testing.T) {
		log.Printf("Running Delete Secret test")
		key := "app/secret/to-delete"
		
		// Set a secret
		log.Printf("Setting secret to delete")
		err := provider.SetSecret(ctx, key, "delete-me", nil)
		if err != nil {
			t.Fatalf("Failed to set secret: %v", err)
		}

		// Verify it exists
		log.Printf("Verifying secret exists")
		_, err = provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get secret before deletion: %v", err)
		}

		// Delete it
		log.Printf("Deleting secret")
		err = provider.DeleteSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to delete secret: %v", err)
		}

		// Verify it's gone
		log.Printf("Verifying secret is deleted")
		_, err = provider.GetSecret(ctx, key)
		assert.Error(t, err)
		log.Printf("Successfully verified deletion")
	})

	t.Run("List Secrets", func(t *testing.T) {
		log.Printf("Running List Secrets test")
		// Create multiple secrets
		secrets := map[string]string{
			"app1/secret1": "value1",
			"app1/secret2": "value2",
			"app2/secret1": "value3",
		}

		// Cleanup function
		defer func() {
			for k := range secrets {
				provider.DeleteSecret(ctx, k)
			}
		}()

		for k, v := range secrets {
			log.Printf("Setting test secret: %s", k)
			err := provider.SetSecret(ctx, k, v, nil)
			if err != nil {
				t.Fatalf("Failed to set test secret %s: %v", k, err)
			}
		}

		// List secrets with prefix app1/
		log.Printf("Listing secrets with prefix app1/")
		list, err := provider.ListSecrets(ctx, "app1/")
		if err != nil {
			t.Fatalf("Failed to list secrets: %v", err)
		}
		assert.Len(t, list, 2)
		assert.Contains(t, list, "app1/secret1")
		assert.Contains(t, list, "app1/secret2")
		assert.NotContains(t, list, "app2/secret1")
		log.Printf("Successfully verified filtered list")

		// List all secrets
		log.Printf("Listing all secrets")
		list, err = provider.ListSecrets(ctx, "")
		if err != nil {
			t.Fatalf("Failed to list all secrets: %v", err)
		}
		assert.Len(t, list, len(secrets))
		log.Printf("Successfully verified full list")
	})

	t.Run("Secret Rotation", func(t *testing.T) {
		log.Printf("Running Secret Rotation test")
		key := "app/rotating-secret"

		// Set initial secret
		log.Printf("Setting initial secret for rotation")
		err := provider.SetSecret(ctx, key, "initial-value", nil)
		if err != nil {
			t.Fatalf("Failed to set initial secret: %v", err)
		}

		// Cleanup
		defer provider.DeleteSecret(ctx, key)

		// Set rotation policy
		log.Printf("Setting rotation policy")
		now := time.Now()
		policy := &providers.RotationPolicy{
			Interval:     24 * time.Hour,
			Length:       32,
			CharacterSet: "alphanumeric",
			LastRotation: now.Add(-48 * time.Hour), // Last rotation was 2 days ago
			NextRotation: now.Add(-24 * time.Hour), // Next rotation was due yesterday
		}

		err = provider.SetRotationPolicy(ctx, key, policy)
		if err != nil {
			t.Fatalf("Failed to set rotation policy: %v", err)
		}

		// Get the policy back
		log.Printf("Getting rotation policy")
		retrievedPolicy, err := provider.GetRotationPolicy(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get rotation policy: %v", err)
		}
		assert.Equal(t, policy.Interval, retrievedPolicy.Interval)
		assert.Equal(t, policy.CharacterSet, retrievedPolicy.CharacterSet)
		assert.Equal(t, policy.Length, retrievedPolicy.Length)
		assert.Equal(t, policy.LastRotation.Unix(), retrievedPolicy.LastRotation.Unix())
		assert.Equal(t, policy.NextRotation.Unix(), retrievedPolicy.NextRotation.Unix())

		// Rotate the secret
		log.Printf("Rotating secret")
		err = provider.RotateSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to rotate secret: %v", err)
		}

		// Get the rotated secret
		log.Printf("Getting rotated secret")
		rotated, err := provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get rotated secret: %v", err)
		}
		assert.NotEqual(t, "initial-value", rotated.Value)
		assert.Equal(t, 32, len(rotated.Value))
		log.Printf("Successfully verified rotation")
	})
}

func TestLocalProvider_ErrorCases(t *testing.T) {
	log.Printf("Starting TestLocalProvider_ErrorCases")
	
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "keeper-local-test")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Created temp directory: %s", tempDir)
	defer func() {
		log.Printf("Cleaning up temp directory: %s", tempDir)
		os.RemoveAll(tempDir)
	}()

	// Create provider
	secretsDir := filepath.Join(tempDir, "secrets")
	log.Printf("Creating provider with secrets directory: %s", secretsDir)
	provider, err := New(secretsDir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("Get Non-existent Secret", func(t *testing.T) {
		log.Printf("Testing get of non-existent secret")
		_, err := provider.GetSecret(ctx, "non/existent/secret")
		assert.Error(t, err)
		log.Printf("Got expected error: %v", err)
	})

	t.Run("Delete Non-existent Secret", func(t *testing.T) {
		log.Printf("Testing delete of non-existent secret")
		err := provider.DeleteSecret(ctx, "non/existent/secret")
		assert.Error(t, err)
		log.Printf("Got expected error: %v", err)
	})

	t.Run("Invalid Secret Path", func(t *testing.T) {
		log.Printf("Testing invalid secret path")
		err := provider.SetSecret(ctx, "", "value", nil)
		assert.Error(t, err)
		log.Printf("Got expected error: %v", err)
	})

	t.Run("List Non-existent Directory", func(t *testing.T) {
		log.Printf("Testing list of non-existent directory")
		list, err := provider.ListSecrets(ctx, "non/existent/path/")
		require.NoError(t, err)
		assert.Empty(t, list)
		log.Printf("Successfully verified empty list")
	})
}

func TestLocalProvider_AdvancedRotation(t *testing.T) {
	log.Printf("Starting TestLocalProvider_AdvancedRotation")

	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "keeper-local-test")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Created temp directory: %s", tempDir)
	defer func() {
		log.Printf("Cleaning up temp directory: %s", tempDir)
		os.RemoveAll(tempDir)
	}()

	// Create provider
	secretsDir := filepath.Join(tempDir, "secrets")
	log.Printf("Creating provider with secrets directory: %s", secretsDir)
	provider, err := New(secretsDir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("Custom Generator", func(t *testing.T) {
		key := "app/custom/secret"
		log.Printf("Testing rotation with custom generator")

		// Set initial secret
		err := provider.SetSecret(ctx, key, "initial-value", nil)
		if err != nil {
			t.Fatalf("Failed to set initial secret: %v", err)
		}

		// Cleanup
		defer provider.DeleteSecret(ctx, key)

		// Set rotation policy with custom generator
		now := time.Now()
		policy := &providers.RotationPolicy{
			Interval:     24 * time.Hour,
			LastRotation: now.Add(-48 * time.Hour), // Last rotation was 2 days ago
			NextRotation: now.Add(-24 * time.Hour), // Next rotation was due yesterday
			CustomGenerator: func() (string, error) {
				return "custom-generated-value", nil
			},
		}

		err = provider.SetRotationPolicy(ctx, key, policy)
		if err != nil {
			t.Fatalf("Failed to set rotation policy: %v", err)
		}

		// Rotate the secret
		err = provider.RotateSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to rotate secret: %v", err)
		}

		// Get the rotated secret
		rotated, err := provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get rotated secret: %v", err)
		}
		assert.Equal(t, "custom-generated-value", rotated.Value)
		assert.Equal(t, "initial-value", rotated.Metadata["previous_value"])
	})

	t.Run("Not Time to Rotate", func(t *testing.T) {
		key := "app/not/time"
		log.Printf("Testing rotation when not yet time")

		// Set initial secret
		err := provider.SetSecret(ctx, key, "initial-value", nil)
		if err != nil {
			t.Fatalf("Failed to set initial secret: %v", err)
		}

		// Cleanup
		defer provider.DeleteSecret(ctx, key)

		// Set rotation policy with future rotation time
		now := time.Now()
		policy := &providers.RotationPolicy{
			Interval:     24 * time.Hour,
			Length:       32,
			CharacterSet: "alphanumeric",
			LastRotation: now,                   // Last rotation was now
			NextRotation: now.Add(24 * time.Hour), // Next rotation is tomorrow
		}

		err = provider.SetRotationPolicy(ctx, key, policy)
		if err != nil {
			t.Fatalf("Failed to set rotation policy: %v", err)
		}

		// Attempt to rotate
		err = provider.RotateSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to call rotate: %v", err)
		}

		// Verify secret was not rotated
		secret, err := provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get secret: %v", err)
		}
		assert.Equal(t, "initial-value", secret.Value)
		assert.Empty(t, secret.Metadata["previous_value"])
	})

	t.Run("Different Character Sets", func(t *testing.T) {
		testCases := []struct {
			name        string
			charSet     string
			validateFn  func(string) bool
		}{
			{
				name:    "numeric",
				charSet: "numeric",
				validateFn: func(s string) bool {
					for _, c := range s {
						if c < '0' || c > '9' {
							return false
						}
					}
					return true
				},
			},
			{
				name:    "hex",
				charSet: "hex",
				validateFn: func(s string) bool {
					for _, c := range s {
						if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
							return false
						}
					}
					return true
				},
			},
			{
				name:    "ascii",
				charSet: "ascii",
				validateFn: func(s string) bool {
					for _, c := range s {
						if c < '!' || c > '~' {
							return false
						}
					}
					return true
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				key := "app/charset/" + tc.name
				log.Printf("Testing rotation with %s character set", tc.name)

				// Set initial secret
				err := provider.SetSecret(ctx, key, "initial-value", nil)
				if err != nil {
					t.Fatalf("Failed to set initial secret: %v", err)
				}

				// Cleanup
				defer provider.DeleteSecret(ctx, key)

				// Set rotation policy
				now := time.Now()
				policy := &providers.RotationPolicy{
					Interval:     24 * time.Hour,
					Length:       32,
					CharacterSet: tc.charSet,
					LastRotation: now.Add(-48 * time.Hour), // Last rotation was 2 days ago
					NextRotation: now.Add(-24 * time.Hour), // Next rotation was due yesterday
				}

				err = provider.SetRotationPolicy(ctx, key, policy)
				if err != nil {
					t.Fatalf("Failed to set rotation policy: %v", err)
				}

				// Rotate the secret
				err = provider.RotateSecret(ctx, key)
				if err != nil {
					t.Fatalf("Failed to rotate secret: %v", err)
				}

				// Get the rotated secret
				rotated, err := provider.GetSecret(ctx, key)
				if err != nil {
					t.Fatalf("Failed to get rotated secret: %v", err)
				}

				assert.NotEqual(t, "initial-value", rotated.Value)
				assert.Equal(t, 32, len(rotated.Value))
				assert.True(t, tc.validateFn(rotated.Value))
			})
		}
	})

	t.Run("Invalid Character Set", func(t *testing.T) {
		key := "app/invalid/charset"
		log.Printf("Testing rotation with invalid character set")

		// Set initial secret
		err := provider.SetSecret(ctx, key, "initial-value", nil)
		if err != nil {
			t.Fatalf("Failed to set initial secret: %v", err)
		}

		// Cleanup
		defer provider.DeleteSecret(ctx, key)

		// Set rotation policy with invalid character set
		now := time.Now()
		policy := &providers.RotationPolicy{
			Interval:     24 * time.Hour,
			Length:       32,
			CharacterSet: "invalid",
			LastRotation: now.Add(-48 * time.Hour), // Last rotation was 2 days ago
			NextRotation: now.Add(-24 * time.Hour), // Next rotation was due yesterday
		}

		err = provider.SetRotationPolicy(ctx, key, policy)
		if err != nil {
			t.Fatalf("Failed to set rotation policy: %v", err)
		}

		// Rotate the secret
		err = provider.RotateSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to rotate secret: %v", err)
		}

		// Get the rotated secret
		rotated, err := provider.GetSecret(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get rotated secret: %v", err)
		}

		// Should use alphanumeric as default
		assert.NotEqual(t, "initial-value", rotated.Value)
		assert.Equal(t, 32, len(rotated.Value))
		for _, c := range rotated.Value {
			assert.True(t, (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'))
		}
	})
}
