package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a secret by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		secret, err := provider.GetSecret(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		fmt.Println(secret.Value)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
