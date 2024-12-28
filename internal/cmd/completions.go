package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// addCompletionFlags adds completion flags to the command
func addCompletionFlags(cmd *cobra.Command) {
	// Register custom completion functions
	if cmd.Name() == "set" || cmd.Name() == "get" || cmd.Name() == "delete" {
		if err := cmd.RegisterFlagCompletionFunc("provider", completeProvider); err != nil {
			panic(err)
		}
	}
}

// completeProvider provides completion for provider types
func completeProvider(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	providers := []string{"local", "vault", "aws"}
	return providers, cobra.ShellCompDirectiveNoFileComp
}

// addDynamicCompletions adds dynamic completions for various commands
func addDynamicCompletions() {
	// Add completion for get command
	if getCmd != nil {
		getCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			
			provider, _ := cmd.Flags().GetString("provider")
			if provider == "" {
				provider = "local" // default provider
			}

			// Get the list of secrets from the current provider
			secrets, err := service.List(cmd.Context(), "")
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			// Filter secrets based on the completion prefix
			var filtered []string
			for _, secret := range secrets {
				if strings.HasPrefix(secret, toComplete) {
					filtered = append(filtered, secret)
				}
			}

			return filtered, cobra.ShellCompDirectiveNoFileComp
		}
	}

	// Add completion for delete command
	if deleteCmd != nil {
		deleteCmd.ValidArgsFunction = getCmd.ValidArgsFunction
	}
}
