package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	backupFormat      string
	includeVersions   bool
	compressBackup    bool
	encryptBackup     bool
	overwriteExisting bool
	restoreVersions   bool
)

var backupCmd = &cobra.Command{
	Use:   "backup [directory]",
	Short: "Backup all secrets to a directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		// Create backup directory if it doesn't exist
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}

		// Set backup directory in provider config
		if err := provider.SetBackupDir(dir); err != nil {
			return fmt.Errorf("failed to set backup directory: %w", err)
		}

		// Backup secrets
		if err := provider.Backup(cmd.Context()); err != nil {
			return fmt.Errorf("failed to backup secrets: %w", err)
		}

		fmt.Printf("Successfully backed up secrets to %s\n", dir)
		return nil
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore [directory]",
	Short: "Restore secrets from a backup directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		// Check if backup directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("backup directory does not exist: %s", dir)
		}

		// Set backup directory in provider config
		if err := provider.SetBackupDir(dir); err != nil {
			return fmt.Errorf("failed to set backup directory: %w", err)
		}

		// Restore secrets
		if err := provider.Restore(cmd.Context()); err != nil {
			return fmt.Errorf("failed to restore secrets: %w", err)
		}

		fmt.Printf("Successfully restored secrets from %s\n", dir)
		return nil
	},
}

func init() {
	// Backup flags
	backupCmd.Flags().StringVar(&backupFormat, "format", "json", "Backup format (json or yaml)")
	backupCmd.Flags().BoolVar(&includeVersions, "versions", false, "Include version history in backup")
	backupCmd.Flags().BoolVar(&compressBackup, "compress", false, "Compress backup file")
	backupCmd.Flags().BoolVar(&encryptBackup, "encrypt", true, "Encrypt backup file")

	// Restore flags
	restoreCmd.Flags().BoolVar(&overwriteExisting, "overwrite", false, "Overwrite existing secrets")
	restoreCmd.Flags().BoolVar(&restoreVersions, "versions", false, "Restore version history")

	// Add commands
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}
