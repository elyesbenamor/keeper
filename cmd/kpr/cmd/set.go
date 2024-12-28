package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	metadata string
)

func init() {
	setCmd.Flags().StringVar(&metadata, "metadata", "", "Metadata key-value pairs (format: key1=value1,key2=value2)")
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a secret value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := getProvider()
		if err != nil {
			return err
		}
		defer provider.Close()

		// Parse metadata
		metadataMap := make(map[string]string)
		if metadata != "" {
			pairs := strings.Split(metadata, ",")
			for _, pair := range pairs {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) != 2 {
					return fmt.Errorf("invalid metadata format: %s", pair)
				}
				metadataMap[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}

		if err := provider.SetSecret(context.Background(), args[0], args[1], metadataMap); err != nil {
			return fmt.Errorf("failed to set secret: %w", err)
		}

		fmt.Printf("Secret '%s' set successfully\n", args[0])
		return nil
	},
}
