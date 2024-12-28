package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	prefix string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		secrets, err := provider.ListSecrets(context.Background(), prefix)
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, key := range secrets {
			fmt.Println(key)
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&prefix, "prefix", "", "List only secrets with this prefix")
	rootCmd.AddCommand(listCmd)
}
