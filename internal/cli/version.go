package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information (set at build time)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("devserv %s\n", Version)
		if GitCommit != "unknown" {
			fmt.Printf("  commit: %s\n", GitCommit)
		}
		if BuildDate != "unknown" {
			fmt.Printf("  built:  %s\n", BuildDate)
		}
	},
}
