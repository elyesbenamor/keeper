package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := provider.DeleteSecret(cmd.Context(), name); err != nil {
			return fmt.Errorf("failed to delete secret: %w", err)
		}
		fmt.Printf("Successfully deleted secret %s\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
