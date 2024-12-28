package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a secret by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		if err := provider.DeleteSecret(context.Background(), args[0]); err != nil {
			return fmt.Errorf("failed to delete secret: %w", err)
		}

		fmt.Printf("Secret '%s' deleted successfully\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
