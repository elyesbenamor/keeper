package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cleanupCmd)
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up invalid secrets",
	Long:  `Remove any secrets that can't be decrypted or are in an invalid format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		if err := provider.CleanupInvalidSecrets(context.Background()); err != nil {
			return fmt.Errorf("failed to cleanup invalid secrets: %w", err)
		}

		fmt.Println("Successfully cleaned up invalid secrets")
		return nil
	},
}
