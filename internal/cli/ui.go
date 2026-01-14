package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/marcelocajueiro/devserv/internal/process"
	"github.com/marcelocajueiro/devserv/internal/tui"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch interactive dashboard",
	Long: `Launch the interactive TUI dashboard for managing services.

The dashboard provides:
  - Real-time service status monitoring
  - Start/stop/restart controls
  - Log viewing
  - Resource usage display

Keyboard shortcuts:
  s       Start selected service
  x       Stop selected service
  r       Restart selected service
  S       Start all services
  X       Stop all services
  l       View logs for selected service
  j/k     Navigate up/down (or arrow keys)
  ?       Show help
  q       Quit

Examples:
  devserv ui
`,
	RunE: runUI,
}

func runUI(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager, err := process.NewManager(cfg)
	if err != nil {
		return err
	}

	if err := tui.Run(manager); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
