package cmd

import (
	"fmt"
	"strings"

	"github.com/keeper/internal/providers"
	"github.com/spf13/cobra"
)

var (
	tags     []string
	schema   string
	metadata []string
)

var setCmd = &cobra.Command{
	Use:   "set [name] [value]",
	Short: "Set a secret",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		value := args[1]

		// Create secret
		secret := providers.NewSecret(name, value)

		// Add tags
		if len(tags) > 0 {
			secret.Tags = tags
		}

		// Add schema
		if schema != "" {
			secret.Schema = schema
		}

		// Add metadata
		if len(metadata) > 0 {
			secret.Metadata = make(map[string]string)
			for _, m := range metadata {
				parts := strings.SplitN(m, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid metadata format: %s (expected key=value)", m)
				}
				secret.Metadata[parts[0]] = parts[1]
			}
		}

		// Set secret
		if err := provider.SetSecret(cmd.Context(), secret); err != nil {
			return fmt.Errorf("failed to set secret: %w", err)
		}

		fmt.Printf("Successfully set secret %s\n", name)
		return nil
	},
}

func init() {
	setCmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags for the secret (comma-separated)")
	setCmd.Flags().StringVar(&schema, "schema", "", "Schema for the secret")
	setCmd.Flags().StringSliceVar(&metadata, "metadata", nil, "Metadata for the secret (comma-separated key=value pairs)")
	rootCmd.AddCommand(setCmd)
}
