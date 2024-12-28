package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Manage secret metadata",
}

var metadataSetCmd = &cobra.Command{
	Use:   "set [name] [key=value...]",
	Short: "Set metadata for a secret",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		secret, err := provider.GetSecret(cmd.Context(), name)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		// Parse metadata key-value pairs
		metadata := make(map[string]string)
		for _, arg := range args[1:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid metadata format: %s", arg)
			}
			metadata[parts[0]] = parts[1]
		}

		// Update metadata
		if secret.Metadata == nil {
			secret.Metadata = make(map[string]string)
		}
		for k, v := range metadata {
			secret.Metadata[k] = v
		}

		// Save updated secret
		if err := provider.SetSecret(cmd.Context(), secret); err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}

		fmt.Printf("Successfully updated metadata for secret %s\n", name)
		return nil
	},
}

var metadataGetCmd = &cobra.Command{
	Use:   "get [name] [key]",
	Short: "Get metadata for a secret",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		secret, err := provider.GetSecret(cmd.Context(), name)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		// Print specific key if provided
		if len(args) > 1 {
			key := args[1]
			value, ok := secret.Metadata[key]
			if !ok {
				return fmt.Errorf("metadata key not found: %s", key)
			}
			fmt.Printf("%s: %s\n", key, value)
			return nil
		}

		// Print all metadata
		if len(secret.Metadata) == 0 {
			fmt.Printf("No metadata found for secret %s\n", name)
			return nil
		}

		fmt.Printf("Metadata for secret %s:\n", name)
		for k, v := range secret.Metadata {
			fmt.Printf("  %s: %s\n", k, v)
		}

		return nil
	},
}

func init() {
	metadataCmd.AddCommand(metadataSetCmd)
	metadataCmd.AddCommand(metadataGetCmd)
	rootCmd.AddCommand(metadataCmd)
}
