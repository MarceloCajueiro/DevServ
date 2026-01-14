package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/marcelomd/devserv/internal/process"
)

var startCmd = &cobra.Command{
	Use:   "start [service...]",
	Short: "Start services",
	Long: `Start one or more services defined in devserv.toml.

If no service names are provided, starts all services.

Examples:
  devserv start           # Start all services
  devserv start api       # Start only the 'api' service
  devserv start api web   # Start 'api' and 'web' services
`,
	RunE: runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager, err := process.NewManager(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start services
	if len(args) == 0 {
		fmt.Println("Starting all services...")
	} else {
		fmt.Printf("Starting services: %v\n", args)
	}

	if err := manager.Start(ctx, args...); err != nil {
		return err
	}

	// Print status
	for _, status := range manager.AllStatus() {
		if status.State.IsActive() {
			fmt.Printf("  %s %s started (PID: %d)\n",
				status.State.Symbol(), status.Name, status.PID)
		}
	}

	// Wait for signal
	fmt.Println("\nPress Ctrl+C to stop all services")

	// Process events and wait for signal
	for {
		select {
		case sig := <-sigCh:
			fmt.Printf("\nReceived %v, stopping services...\n", sig)
			return manager.Shutdown(ctx)

		case event := <-manager.Events():
			switch event.Type {
			case process.EventCrashed:
				fmt.Printf("  ✖ %s crashed: %v\n", event.Service, event.Data)
			case process.EventStopped:
				fmt.Printf("  ○ %s stopped\n", event.Service)
			}
		}
	}
}
