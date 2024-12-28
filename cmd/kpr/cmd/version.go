package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(listVersionsCmd)
	versionCmd.AddCommand(rollbackCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage secret versions",
	Long:  `Commands for managing secret versions, including listing versions and rolling back to previous versions.`,
}

var listVersionsCmd = &cobra.Command{
	Use:   "list [key]",
	Short: "List all versions of a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		key := args[0]
		versions, err := provider.ListSecretVersions(context.Background(), key)
		if err != nil {
			return fmt.Errorf("failed to list versions: %w", err)
		}

		fmt.Printf("Versions for secret '%s':\n", key)
		for _, v := range versions {
			secret, err := provider.GetSecretVersion(context.Background(), key, v)
			if err != nil {
				fmt.Printf("  Version %d: Error - %v\n", v, err)
				continue
			}
			fmt.Printf("  Version %d:\n", v)
			fmt.Printf("    Created: %s\n", secret.Created)
			fmt.Printf("    Updated: %s\n", secret.Updated)
			if len(secret.Metadata) > 0 {
				fmt.Println("    Metadata:")
				for k, v := range secret.Metadata {
					fmt.Printf("      %s: %s\n", k, v)
				}
			}
		}

		return nil
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [key] [version]",
	Short: "Roll back a secret to a previous version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		key := args[0]
		version, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}

		if err := provider.RollbackSecret(context.Background(), key, version); err != nil {
			return fmt.Errorf("failed to rollback secret: %w", err)
		}

		fmt.Printf("Successfully rolled back secret '%s' to version %d\n", key, version)
		return nil
	},
}
