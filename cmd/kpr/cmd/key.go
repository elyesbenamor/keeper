package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage encryption keys",
	Long:  `Commands for managing encryption keys, including rotation and deletion.`,
}

// keyRotateCmd represents the key rotate command
var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate the master encryption key",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		if err := provider.RotateKey(context.Background()); err != nil {
			return fmt.Errorf("failed to rotate key: %w", err)
		}

		fmt.Println("Master key rotated successfully")
		return nil
	},
}

// keyDeleteCmd represents the key delete command
var deleteKeyCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the master encryption key",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		if err := provider.DeleteKey(context.Background()); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}

		fmt.Println("Master key deleted successfully")
		return nil
	},
}

func init() {
	keyCmd.AddCommand(rotateCmd)
	keyCmd.AddCommand(deleteKeyCmd)
	rootCmd.AddCommand(keyCmd)
}
