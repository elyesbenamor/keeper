package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kpr",
	Short: "Keeper - A secure secret management tool",
	Long: `Keeper is a command-line tool for securely managing secrets.
It supports multiple storage providers and encrypts secrets at rest.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add commands
	AddKeyCommands(rootCmd)
}
