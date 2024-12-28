package cmd

import (
	"fmt"
	"strings"

	"github.com/keeper/internal/providers"
	"github.com/spf13/cobra"
)

var (
	prefix string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [prefix]",
	Short: "List secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		// List all secrets
		secrets, err := provider.ListSecrets(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		// Filter by prefix if provided
		if len(args) > 0 {
			prefix := args[0]
			var filtered []*providers.Secret
			for _, secret := range secrets {
				if strings.HasPrefix(secret.Name, prefix) {
					filtered = append(filtered, secret)
				}
			}
			secrets = filtered
		}

		// Print secrets
		if len(secrets) == 0 {
			fmt.Println("No secrets found")
			return nil
		}

		fmt.Printf("Found %d secrets:\n", len(secrets))
		for _, secret := range secrets {
			fmt.Printf("- %s\n", secret.Name)
			if len(secret.Tags) > 0 {
				fmt.Printf("  Tags: %s\n", strings.Join(secret.Tags, ", "))
			}
			if secret.Schema != "" {
				fmt.Printf("  Schema: %s\n", secret.Schema)
			}
			if len(secret.Metadata) > 0 {
				fmt.Println("  Metadata:")
				for k, v := range secret.Metadata {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Printf("  Created: %s\n", secret.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Updated: %s\n", secret.UpdatedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
