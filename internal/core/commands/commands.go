package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/keeper/internal/config"
	"github.com/keeper/internal/core/service"
	"github.com/keeper/internal/providers"
)

// GetCommand represents the get command
type GetCommand struct {
	cfg *config.Config
}

// NewGetCommand creates a new get command
func NewGetCommand(cfg *config.Config) *GetCommand {
	return &GetCommand{cfg: cfg}
}

// Execute executes the get command
func (c *GetCommand) Execute(ctx context.Context, key string) error {
	// Create service config
	svcConfig := service.Config{
		Type: c.cfg.DefaultProvider,
		Parameters: map[string]interface{}{
			"secretsDir": filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"path":      filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"token":     os.Getenv("VAULT_TOKEN"),
			"address":   os.Getenv("VAULT_ADDR"),
			"region":    os.Getenv("AWS_REGION"),
		},
	}

	// Create service
	svc, err := service.New(svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	// Get secret
	secret, err := svc.GetSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Print secret value
	fmt.Println(secret.Value)
	return nil
}

// SetCommand represents the set command
type SetCommand struct {
	cfg      *config.Config
	metadata map[string]string
}

// NewSetCommand creates a new set command
func NewSetCommand(cfg *config.Config) *SetCommand {
	return &SetCommand{
		cfg:      cfg,
		metadata: make(map[string]string),
	}
}

// Execute executes the set command
func (c *SetCommand) Execute(ctx context.Context, key, value string) error {
	// Create service config
	svcConfig := service.Config{
		Type: c.cfg.DefaultProvider,
		Parameters: map[string]interface{}{
			"secretsDir": filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"path":      filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"token":     os.Getenv("VAULT_TOKEN"),
			"address":   os.Getenv("VAULT_ADDR"),
			"region":    os.Getenv("AWS_REGION"),
		},
	}

	// Create service
	svc, err := service.New(svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	// Set secret
	err = svc.SetSecret(ctx, key, value, c.metadata)
	if err != nil {
		return fmt.Errorf("failed to set secret: %w", err)
	}

	fmt.Printf("Secret '%s' set successfully\n", key)
	return nil
}

// ListCommand represents the list command
type ListCommand struct {
	cfg *config.Config
}

// NewListCommand creates a new list command
func NewListCommand(cfg *config.Config) *ListCommand {
	return &ListCommand{cfg: cfg}
}

// Execute executes the list command
func (c *ListCommand) Execute(ctx context.Context, prefix string) error {
	// Create service config
	svcConfig := service.Config{
		Type: c.cfg.DefaultProvider,
		Parameters: map[string]interface{}{
			"secretsDir": filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"path":      filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"token":     os.Getenv("VAULT_TOKEN"),
			"address":   os.Getenv("VAULT_ADDR"),
			"region":    os.Getenv("AWS_REGION"),
		},
	}

	// Create service
	svc, err := service.New(svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	// List secrets
	secrets, err := svc.ListSecrets(ctx, prefix)
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	// Print secrets
	for _, secret := range secrets {
		fmt.Println(secret)
	}
	return nil
}

// DeleteCommand represents the delete command
type DeleteCommand struct {
	cfg *config.Config
}

// NewDeleteCommand creates a new delete command
func NewDeleteCommand(cfg *config.Config) *DeleteCommand {
	return &DeleteCommand{cfg: cfg}
}

// Execute executes the delete command
func (c *DeleteCommand) Execute(ctx context.Context, key string) error {
	// Create service config
	svcConfig := service.Config{
		Type: c.cfg.DefaultProvider,
		Parameters: map[string]interface{}{
			"secretsDir": filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"path":      filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"token":     os.Getenv("VAULT_TOKEN"),
			"address":   os.Getenv("VAULT_ADDR"),
			"region":    os.Getenv("AWS_REGION"),
		},
	}

	// Create service
	svc, err := service.New(svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	// Delete secret
	err = svc.DeleteSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	fmt.Printf("Secret '%s' deleted successfully\n", key)
	return nil
}

// RotateCommand represents the rotate command
type RotateCommand struct {
	cfg *config.Config
}

// NewRotateCommand creates a new rotate command
func NewRotateCommand(cfg *config.Config) *RotateCommand {
	return &RotateCommand{cfg: cfg}
}

// Execute executes the rotate command
func (c *RotateCommand) Execute(ctx context.Context, key string) error {
	// Create service config
	svcConfig := service.Config{
		Type: c.cfg.DefaultProvider,
		Parameters: map[string]interface{}{
			"secretsDir": filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"path":      filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"token":     os.Getenv("VAULT_TOKEN"),
			"address":   os.Getenv("VAULT_ADDR"),
			"region":    os.Getenv("AWS_REGION"),
		},
	}

	// Create service
	svc, err := service.New(svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	// Rotate secret
	err = svc.RotateSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to rotate secret: %w", err)
	}

	fmt.Printf("Secret '%s' rotated successfully\n", key)
	return nil
}

// SetPolicyCommand represents the set-policy command
type SetPolicyCommand struct {
	cfg *config.Config
}

// NewSetPolicyCommand creates a new set-policy command
func NewSetPolicyCommand(cfg *config.Config) *SetPolicyCommand {
	return &SetPolicyCommand{cfg: cfg}
}

// Execute executes the set-policy command
func (c *SetPolicyCommand) Execute(ctx context.Context, key string, interval time.Duration, length int, characterSet string) error {
	// Create service config
	svcConfig := service.Config{
		Type: c.cfg.DefaultProvider,
		Parameters: map[string]interface{}{
			"secretsDir": filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"path":      filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"token":     os.Getenv("VAULT_TOKEN"),
			"address":   os.Getenv("VAULT_ADDR"),
			"region":    os.Getenv("AWS_REGION"),
		},
	}

	// Create service
	svc, err := service.New(svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	now := time.Now()
	policy := &providers.RotationPolicy{
		Interval:     interval,
		Length:       length,
		CharacterSet: characterSet,
		LastRotation: now,
		NextRotation: now.Add(interval),
	}

	err = svc.SetRotationPolicy(ctx, key, policy)
	if err != nil {
		return fmt.Errorf("failed to set rotation policy: %w", err)
	}

	fmt.Printf("Rotation policy set for '%s'\n", key)
	return nil
}

// MetadataCommand represents the metadata command
type MetadataCommand struct {
	cfg *config.Config
}

// NewMetadataCommand creates a new metadata command
func NewMetadataCommand(cfg *config.Config) *MetadataCommand {
	return &MetadataCommand{cfg: cfg}
}

// Execute executes the metadata command
func (c *MetadataCommand) Execute(ctx context.Context, key string) error {
	// Create service config
	svcConfig := service.Config{
		Type: c.cfg.DefaultProvider,
		Parameters: map[string]interface{}{
			"secretsDir": filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"path":      filepath.Join(c.cfg.DefaultDataDir, "secrets"),
			"token":     os.Getenv("VAULT_TOKEN"),
			"address":   os.Getenv("VAULT_ADDR"),
			"region":    os.Getenv("AWS_REGION"),
		},
	}

	// Create service
	svc, err := service.New(svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	// Get secret metadata
	metadata, err := svc.GetSecretMetadata(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get secret metadata: %w", err)
	}

	// Print metadata in a formatted way
	fmt.Printf("Metadata for '%s':\n", key)
	fmt.Println("User Metadata:")
	for k, v := range metadata.UserMetadata {
		fmt.Printf("  %s: %s\n", k, v)
	}
	
	fmt.Println("\nSystem Metadata:")
	fmt.Printf("  Created: %s\n", metadata.Created.Format(time.RFC3339))
	fmt.Printf("  Last Modified: %s\n", metadata.LastModified.Format(time.RFC3339))
	
	if metadata.RotationPolicy != nil {
		fmt.Println("\nRotation Policy:")
		fmt.Printf("  Interval: %s\n", metadata.RotationPolicy.Interval)
		fmt.Printf("  Length: %d\n", metadata.RotationPolicy.Length)
		fmt.Printf("  Character Set: %s\n", metadata.RotationPolicy.CharacterSet)
		fmt.Printf("  Last Rotation: %s\n", metadata.RotationPolicy.LastRotation.Format(time.RFC3339))
		fmt.Printf("  Next Rotation: %s\n", metadata.RotationPolicy.NextRotation.Format(time.RFC3339))
	}

	if metadata.PreviousVersion != "" {
		fmt.Println("\nPrevious Version Available: Yes")
	}

	return nil
}

// AddCommands adds all the CLI commands to the root command
func AddCommands(rootCmd *cobra.Command) {
	// Add key management commands
	AddKeyCommands(rootCmd)
	
	// Add existing commands
	AddGetCommand(rootCmd)
	AddSetCommand(rootCmd)
	AddListCommand(rootCmd)
	AddDeleteCommand(rootCmd)
}

// AddGetCommand adds the get command to the root command
func AddGetCommand(rootCmd *cobra.Command) {
	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			return NewGetCommand(cfg).Execute(cmd.Context(), args[0])
		},
	}
	rootCmd.AddCommand(getCmd)
}

// AddSetCommand adds the set command to the root command
func AddSetCommand(rootCmd *cobra.Command) {
	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			setCommand := NewSetCommand(cfg)
			metadata, err := cmd.Flags().GetString("metadata")
			if err != nil {
				return err
			}
			if metadata != "" {
				// Parse metadata string in format "key1=value1,key2=value2"
				pairs := strings.Split(metadata, ",")
				for _, pair := range pairs {
					kv := strings.Split(pair, "=")
					if len(kv) != 2 {
						return fmt.Errorf("invalid metadata format: %s", pair)
					}
					setCommand.metadata[kv[0]] = kv[1]
				}
			}
			return setCommand.Execute(cmd.Context(), args[0], args[1])
		},
	}
	setCmd.Flags().String("metadata", "", "Metadata in format 'key1=value1,key2=value2'")
	rootCmd.AddCommand(setCmd)
}

// AddListCommand adds the list command to the root command
func AddListCommand(rootCmd *cobra.Command) {
	listCmd := &cobra.Command{
		Use:   "list [prefix]",
		Short: "List secrets under a path",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			listCmd := NewListCommand(cfg)
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return listCmd.Execute(cmd.Context(), path)
		},
	}
	rootCmd.AddCommand(listCmd)
}

// AddDeleteCommand adds the delete command to the root command
func AddDeleteCommand(rootCmd *cobra.Command) {
	deleteCmd := &cobra.Command{
		Use:   "delete [key]",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			deleteCmd := NewDeleteCommand(cfg)
			return deleteCmd.Execute(cmd.Context(), args[0])
		},
	}
	rootCmd.AddCommand(deleteCmd)
}
