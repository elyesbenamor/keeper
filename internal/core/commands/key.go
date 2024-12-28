package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/keeper/internal/keychain"
)

// AddKeyCommands adds key management commands to the root command
func AddKeyCommands(root *cobra.Command) {
	keyCmd := &cobra.Command{
		Use:   "key",
		Short: "Manage the master encryption key",
		Long: `Manage the master encryption key used to encrypt secrets.
This key is stored in your system's secure keychain.`,
	}

	rotateCmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate the master encryption key",
		Long: `Generate a new master encryption key and re-encrypt all secrets.
This operation may take some time depending on the number of secrets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			appDir := filepath.Join(homeDir, ".keeper")

			km := keychain.New(appDir)
			if err := km.RotateKey(); err != nil {
				return fmt.Errorf("failed to rotate key: %w", err)
			}

			fmt.Println("Master key rotated successfully")
			return nil
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the master encryption key",
		Long: `Delete the master encryption key from the system keychain.
WARNING: This will make all encrypted secrets unreadable!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			appDir := filepath.Join(homeDir, ".keeper")

			km := keychain.New(appDir)
			if err := km.DeleteKey(); err != nil {
				return fmt.Errorf("failed to delete key: %w", err)
			}

			fmt.Println("Master key deleted successfully")
			return nil
		},
	}

	keyCmd.AddCommand(rotateCmd)
	keyCmd.AddCommand(deleteCmd)
	root.AddCommand(keyCmd)
}
