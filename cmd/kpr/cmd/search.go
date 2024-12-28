package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/keeper/internal/providers"
	"github.com/spf13/cobra"
)

var (
	searchTags      []string
	searchSchema    string
	createdAfter   string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse search options
		opts := providers.SearchOptions{
			Tags:   searchTags,
			Schema: searchSchema,
		}

		if createdAfter != "" {
			t, err := time.Parse("2006-01-02", createdAfter)
			if err != nil {
				return fmt.Errorf("invalid date format for created-after: %w", err)
			}
			opts.CreatedAfter = t
		}

		// Search for secrets
		secrets, err := provider.SearchSecrets(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search secrets: %w", err)
		}

		// Print results
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
	searchCmd.Flags().StringSliceVar(&searchTags, "tags", nil, "Filter by tags (comma-separated)")
	searchCmd.Flags().StringVar(&searchSchema, "schema", "", "Filter by schema")
	searchCmd.Flags().StringVar(&createdAfter, "created-after", "", "Filter by creation date (YYYY-MM-DD)")
	rootCmd.AddCommand(searchCmd)
}
