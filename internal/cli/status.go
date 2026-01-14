package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/marcelomd/devserv/internal/process"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service status",
	Long: `Display the status of all configured services.

Shows each service's state, PID, port, and uptime.

Examples:
  devserv status
`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager, err := process.NewManager(cfg)
	if err != nil {
		return err
	}

	statuses := manager.AllStatus()

	if len(statuses) == 0 {
		fmt.Println("No services configured")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tSTATUS\tPID\tPORT\tUPTIME")
	fmt.Fprintln(w, "-------\t------\t---\t----\t------")

	for _, s := range statuses {
		pid := "-"
		if s.PID > 0 {
			pid = fmt.Sprintf("%d", s.PID)
		}

		port := "-"
		if s.Port > 0 {
			port = fmt.Sprintf("%d", s.Port)
		}

		uptime := "-"
		if s.Uptime > 0 {
			uptime = formatDuration(s.Uptime)
		}

		statusStr := fmt.Sprintf("%s %s", s.State.Symbol(), s.State.String())

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			s.Name, statusStr, pid, port, uptime)
	}

	w.Flush()
	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
