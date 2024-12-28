package cmd

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get build info
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return fmt.Errorf("failed to read build info")
		}

		// Print version info
		fmt.Printf("Keeper CLI Version: %s\n", info.Main.Version)
		fmt.Printf("Go Version: %s\n", info.GoVersion)
		fmt.Printf("Build Time: %s\n", time.Now().Format(time.RFC3339))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
