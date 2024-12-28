package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up expired secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get secrets that have expired
		secrets, err := provider.ListSecrets(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		// Count expired secrets
		expired := 0
		for _, secret := range secrets {
			if secret.UpdatedAt.Add(24 * time.Hour).Before(time.Now()) {
				if err := provider.DeleteSecret(cmd.Context(), secret.Name); err != nil {
					return fmt.Errorf("failed to delete expired secret %s: %w", secret.Name, err)
				}
				expired++
			}
		}

		fmt.Printf("Successfully cleaned up %d expired secrets\n", expired)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}
