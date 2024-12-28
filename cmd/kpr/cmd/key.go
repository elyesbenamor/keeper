package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage encryption keys",
}

// keyRotateCmd represents the key rotate command
var keyRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate encryption key",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Key rotation not implemented yet")
		return nil
	},
}

// keyDeleteCmd represents the key delete command
var keyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete encryption key",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Key deletion not implemented yet")
		return nil
	},
}

func init() {
	keyCmd.AddCommand(keyRotateCmd)
	keyCmd.AddCommand(keyDeleteCmd)
	rootCmd.AddCommand(keyCmd)
}
