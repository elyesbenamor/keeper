package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/keeper/internal/keychain"
	"github.com/keeper/internal/providers"
	"github.com/keeper/internal/providers/local"
	"github.com/spf13/cobra"
)

var (
	configDir string
	provider  providers.Provider
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpr",
	Short: "A secure secret management tool",
	Long: `Keeper is a command-line tool for managing secrets securely.
It supports storing secrets with metadata, schemas, and versioning.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Initialize keychain
		kc, err := keychain.New()
		if err != nil {
			return fmt.Errorf("failed to initialize keychain: %w", err)
		}

		// Initialize provider
		p, err := local.New(configDir, kc)
		if err != nil {
			return fmt.Errorf("failed to initialize provider: %w", err)
		}

		// Initialize provider
		if err := p.Initialize(cmd.Context()); err != nil {
			return fmt.Errorf("failed to initialize provider: %w", err)
		}

		provider = p
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVar(&configDir, "config", filepath.Join(home, ".keeper"), "config directory")
}
