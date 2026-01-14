package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/marcelocajueiro/devserv/internal/logs"
	"github.com/marcelocajueiro/devserv/internal/process"
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View service logs",
	Long: `View logs for a specific service.

Use -f to follow (tail) logs in real-time.

Examples:
  devserv logs api         # Show recent logs for 'api'
  devserv logs api -f      # Follow logs for 'api'
  devserv logs api -n 50   # Show last 50 lines
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogs,
}

var (
	followLogs bool
	numLines   int
)

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVarP(&numLines, "lines", "n", 100, "Number of lines to show")
}

func runLogs(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager, err := process.NewManager(cfg)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		// List available services
		fmt.Println("Available services:")
		for _, name := range cfg.ServiceNames() {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println("\nUsage: devserv logs <service> [-f]")
		return nil
	}

	serviceName := args[0]

	// Verify service exists
	if cfg.GetService(serviceName) == nil {
		return fmt.Errorf("service not found: %s", serviceName)
	}

	// Get latest log file
	logFile, err := manager.LogManager().GetLatestLog(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	if logFile == nil {
		fmt.Printf("No logs found for service: %s\n", serviceName)
		return nil
	}

	// Create reader
	opts := logs.ReaderOptions{
		Follow: followLogs,
		Lines:  numLines,
	}

	reader, err := manager.LogManager().CreateReader(logFile.Path, opts)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	fmt.Printf("Logs for %s (%s)\n", serviceName, logFile.Path)
	if followLogs {
		fmt.Println("Press Ctrl+C to stop following")
	}
	fmt.Println()

	// Read and print entries
	for entry := range reader.Read(ctx) {
		printLogEntry(entry)
	}

	return nil
}

func printLogEntry(entry logs.Entry) {
	timestamp := entry.Timestamp.Format("15:04:05.000")

	// Color based on stream
	streamIndicator := "│"
	if entry.Stream == "stderr" {
		streamIndicator = "║"
	}

	fmt.Printf("%s %s %s\n", timestamp, streamIndicator, entry.Message)
}
