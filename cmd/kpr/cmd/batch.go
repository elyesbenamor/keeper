package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/keeper/internal/providers"
	"github.com/spf13/cobra"
)

var batchSetCmd = &cobra.Command{
	Use:   "batch-set [file]",
	Short: "Set multiple secrets from a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		// Read JSON file
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Parse secrets
		var secrets []*providers.Secret
		if err := json.Unmarshal(data, &secrets); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Set each secret
		for _, secret := range secrets {
			if err := provider.SetSecret(cmd.Context(), secret); err != nil {
				return fmt.Errorf("failed to set secret %s: %w", secret.Name, err)
			}
		}

		fmt.Printf("Successfully set %d secrets\n", len(secrets))
		return nil
	},
}

var batchGetCmd = &cobra.Command{
	Use:   "batch-get [file]",
	Short: "Get multiple secrets and save to a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		// List all secrets
		secrets, err := provider.ListSecrets(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		// Marshal to JSON
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal secrets: %w", err)
		}

		// Write to file
		if err := ioutil.WriteFile(path, data, 0600); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Successfully exported %d secrets to %s\n", len(secrets), path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(batchSetCmd)
	rootCmd.AddCommand(batchGetCmd)
}
