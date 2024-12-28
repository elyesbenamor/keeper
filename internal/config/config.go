package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config represents the main configuration structure
type Config struct {
	DefaultProvider string                       `yaml:"default_provider"`
	Providers       map[string]ProviderConfig    `yaml:"providers"`
	Encryption      EncryptionConfig            `yaml:"encryption"`
	DefaultDataDir  string                      `yaml:"default_data_dir"`
}

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	Type       string                 `yaml:"type"`
	Parameters map[string]interface{} `yaml:"parameters"`
}

// EncryptionConfig holds local encryption settings
type EncryptionConfig struct {
	Algorithm string `yaml:"algorithm"`
	KeyFile   string `yaml:"key_file"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	defaultDataDir := filepath.Join(home, ".keeper")
	return &Config{
		DefaultProvider: "local",
		DefaultDataDir:  defaultDataDir,
		Providers: map[string]ProviderConfig{
			"local": {
				Type: "local",
				Parameters: map[string]interface{}{
					"path": filepath.Join(defaultDataDir, "secrets"),
				},
			},
		},
		Encryption: EncryptionConfig{
			Algorithm: "aes-256-gcm",
			KeyFile:   filepath.Join(defaultDataDir, "master.key"),
		},
	}, nil
}

// Load reads configuration from the specified file
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %w", err)
		}
		configPath = filepath.Join(home, ".keeper", "config.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig()
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	// Set default data directory if not specified
	if config.DefaultDataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %w", err)
		}
		config.DefaultDataDir = filepath.Join(home, ".keeper")
	}

	return &config, nil
}
