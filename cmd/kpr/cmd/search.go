package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/keeper/internal/providers"
	"github.com/spf13/cobra"
)

var (
	searchMetadata string
	sortBy        string
	sortDesc      bool
	maxResults    int
)

func init() {
	searchCmd.Flags().StringVar(&searchMetadata, "metadata", "", "Metadata key-value pairs to search for (format: key1=value1,key2=value2)")
	searchCmd.Flags().StringVar(&sortBy, "sort-by", "", "Sort results by field (created, updated)")
	searchCmd.Flags().BoolVar(&sortDesc, "desc", false, "Sort in descending order")
	searchCmd.Flags().IntVar(&maxResults, "max", 0, "Maximum number of results to return (0 for unlimited)")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for secrets based on metadata",
	Long: `Search for secrets based on metadata key-value pairs.
Example: kpr search --metadata "environment=prod,service=api"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		// Parse metadata
		metadata := make(map[string]string)
		if searchMetadata != "" {
			pairs := strings.Split(searchMetadata, ",")
			for _, pair := range pairs {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) != 2 {
					return fmt.Errorf("invalid metadata format: %s", pair)
				}
				metadata[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}

		// Search for secrets
		secrets, err := provider.SearchSecrets(context.Background(), metadata)
		if err != nil {
			return fmt.Errorf("failed to search secrets: %w", err)
		}

		// Sort results if requested
		if sortBy != "" {
			sortSecrets(secrets, sortBy, sortDesc)
		}

		// Limit results if requested
		if maxResults > 0 && len(secrets) > maxResults {
			secrets = secrets[:maxResults]
		}

		// Print results
		for _, secret := range secrets {
			fmt.Printf("Found matching secret:\n")
			fmt.Printf("  Value: %s\n", secret.Value)
			if len(secret.Metadata) > 0 {
				fmt.Println("  Metadata:")
				for k, v := range secret.Metadata {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Printf("  Version: %d\n", secret.Version)
			fmt.Printf("  Created: %s\n", secret.Created)
			fmt.Printf("  Updated: %s\n", secret.Updated)
			fmt.Println()
		}

		return nil
	},
}

func sortSecrets(secrets []*providers.Secret, sortBy string, sortDesc bool) {
	switch sortBy {
	case "created":
		if sortDesc {
			sort.Slice(secrets, func(i, j int) bool {
				return secrets[i].Created.After(secrets[j].Created)
			})
		} else {
			sort.Slice(secrets, func(i, j int) bool {
				return secrets[i].Created.Before(secrets[j].Created)
			})
		}
	case "updated":
		if sortDesc {
			sort.Slice(secrets, func(i, j int) bool {
				return secrets[i].Updated.After(secrets[j].Updated)
			})
		} else {
			sort.Slice(secrets, func(i, j int) bool {
				return secrets[i].Updated.Before(secrets[j].Updated)
			})
		}
	}
}
