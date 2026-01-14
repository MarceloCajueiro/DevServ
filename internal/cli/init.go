package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new devserv configuration",
	Long: `Create a new devserv.toml configuration file in the current directory.

The generated file includes example service definitions that you can
customize for your project.

Examples:
  devserv init
`,
	RunE: runInit,
}

const exampleConfig = `# DevServ Configuration
# https://github.com/marcelomd/devserv

# Define your services below. Each service needs at least a name and command.
# Optional: directory (working directory), port (for port conflict detection)

[[services]]
name = "api"
command = "go run ./cmd/api"
directory = "./backend"
port = 8080

[[services]]
name = "frontend"
command = "npm run dev"
directory = "./frontend"
port = 3000

[[services]]
name = "worker"
command = "python worker.py"
directory = "./worker"

# Database example (using Docker)
# [[services]]
# name = "postgres"
# command = "docker compose up postgres"
# port = 5432

# Redis example
# [[services]]
# name = "redis"
# command = "redis-server"
# port = 6379
`

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "devserv.toml"

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	// Write config file
	if err := os.WriteFile(configPath, []byte(exampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	absPath, _ := filepath.Abs(configPath)
	fmt.Printf("Created %s\n", absPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit devserv.toml to configure your services")
	fmt.Println("  2. Run 'devserv start' to start all services")
	fmt.Println("  3. Run 'devserv ui' for the interactive dashboard")

	return nil
}
