package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MarceloCajueiro/DevServ/internal/config"
	"github.com/MarceloCajueiro/DevServ/internal/process"
)

var restartCmd = &cobra.Command{
	Use:   "restart [service...]",
	Short: "Restart services",
	Long: `Restart one or more services.

If no service names are provided, restarts all services.

Examples:
  devserv restart           # Restart all services
  devserv restart api       # Restart only the 'api' service
  devserv restart api web   # Restart 'api' and 'web' services
`,
	RunE: runRestart,
}

func runRestart(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager, err := process.NewManager(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultShutdownTimeout*2)
	defer cancel()

	if len(args) == 0 {
		fmt.Println("Restarting all services...")
	} else {
		fmt.Printf("Restarting services: %v\n", args)
	}

	if err := manager.Restart(ctx, args...); err != nil {
		return err
	}

	// Print status
	for _, status := range manager.AllStatus() {
		if status.State.IsActive() {
			fmt.Printf("  %s %s restarted (PID: %d)\n",
				status.State.Symbol(), status.Name, status.PID)
		}
	}

	return nil
}
