package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		secret, err := provider.GetSecret(cmd.Context(), name)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		// Print secret details
		fmt.Printf("Name: %s\n", secret.Name)
		fmt.Printf("Value: %s\n", secret.Value)
		if len(secret.Tags) > 0 {
			fmt.Printf("Tags: %v\n", secret.Tags)
		}
		if secret.Schema != "" {
			fmt.Printf("Schema: %s\n", secret.Schema)
		}
		if len(secret.Metadata) > 0 {
			fmt.Println("Metadata:")
			for k, v := range secret.Metadata {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		fmt.Printf("Created: %s\n", secret.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", secret.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
