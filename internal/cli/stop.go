package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/MarceloCajueiro/DevServ/internal/config"
	"github.com/MarceloCajueiro/DevServ/internal/process"
)

var stopCmd = &cobra.Command{
	Use:   "stop [service...]",
	Short: "Stop services",
	Long: `Stop one or more running services.

If no service names are provided, stops all services.

Examples:
  devserv stop           # Stop all services
  devserv stop api       # Stop only the 'api' service
  devserv stop api web   # Stop 'api' and 'web' services
`,
	RunE: runStop,
}

var forceStop bool

func init() {
	stopCmd.Flags().BoolVarP(&forceStop, "force", "f", false, "Force kill services (SIGKILL)")
}

func runStop(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager, err := process.NewManager(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultShutdownTimeout)
	defer cancel()

	if len(args) == 0 {
		fmt.Println("Stopping all services...")
	} else {
		fmt.Printf("Stopping services: %v\n", args)
	}

	if forceStop {
		// Force kill
		names := args
		if len(names) == 0 {
			names = manager.ServiceNames()
		}
		for _, name := range names {
			if err := manager.Kill(name); err != nil {
				fmt.Printf("  ✖ Failed to kill %s: %v\n", name, err)
			} else {
				fmt.Printf("  ✖ %s killed\n", name)
			}
		}
		return nil
	}

	// Graceful stop
	start := time.Now()
	if err := manager.Stop(ctx, args...); err != nil {
		return err
	}

	fmt.Printf("All services stopped in %v\n", time.Since(start).Round(time.Millisecond))
	return nil
}
