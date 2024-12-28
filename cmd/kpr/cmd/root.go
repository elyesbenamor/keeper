package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/keeper/internal/providers"
	"github.com/keeper/internal/providers/local"
)

var (
	cfgFile  string
	provider string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpr",
	Short: "A secure secret manager",
	Long: `Keeper is a secure secret manager that helps you store and manage
sensitive information like API keys, passwords, and other secrets.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file (default is $HOME/.keeper/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "local", "Secret provider to use (local, vault, aws, azure)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".keeper" (without extension).
		viper.AddConfigPath(filepath.Join(home, ".keeper"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// getProvider returns the configured secret provider
func getProvider() (providers.Provider, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	switch provider {
	case "local":
		provider, err := local.New(filepath.Join(home, ".keeper"))
		if err != nil {
			return nil, fmt.Errorf("failed to create local provider: %w", err)
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// parseMetadata parses metadata string into a map
func parseMetadata(metadata string) map[string]string {
	result := make(map[string]string)
	if metadata == "" {
		return result
	}

	pairs := strings.Split(metadata, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return result
}
