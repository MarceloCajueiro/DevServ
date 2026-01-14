// Package cli implements the command-line interface for devserv.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/marcelocajueiro/devserv/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "devserv",
	Short: "Local development server manager",
	Long: `DevServ is a CLI tool for managing multiple local development services.

It provides process management, log aggregation, and an interactive
TUI dashboard for monitoring your development stack.

Example configuration (devserv.toml):

  [[services]]
  name = "api"
  command = "go run ./cmd/api"
  port = 8080

  [[services]]
  name = "frontend"
  command = "npm run dev"
  port = 3000
`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file path (default: ./devserv.toml, fallback: ~/.devserv/config.toml)")

	// Add subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(uiCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	// Skip config loading for commands that don't need it
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init", "version", "help", "--help", "-h":
			return
		}
	}
}

func loadConfig() (*config.Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
