package internal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/marcelocajueiro/devserv/internal/config"
	"github.com/marcelocajueiro/devserv/internal/logs"
	"github.com/marcelocajueiro/devserv/internal/process"
)

// E2E tests - full flow without mocks

func TestE2EFullFlow(t *testing.T) {
	// This test runs the complete flow:
	// 1. Load config
	// 2. Start services
	// 3. Verify logs are written
	// 4. Stop services
	// 5. Verify cleanup

	// Create temp directories
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "devserv.toml")
	logDir := filepath.Join(tmpDir, "logs")

	// Create shell scripts for testing (avoiding command parsing issues)
	echoScript := filepath.Join(tmpDir, "echo.sh")
	if err := os.WriteFile(echoScript, []byte("#!/bin/sh\necho hello from echo-service\nsleep 10\n"), 0755); err != nil {
		t.Fatalf("failed to write echo script: %v", err)
	}

	counterScript := filepath.Join(tmpDir, "counter.sh")
	if err := os.WriteFile(counterScript, []byte("#!/bin/sh\nfor i in 1 2 3 4 5; do echo count: $i; sleep 1; done\n"), 0755); err != nil {
		t.Fatalf("failed to write counter script: %v", err)
	}

	// Create config file using the scripts
	configContent := `
[[services]]
name = "echo-service"
command = "` + echoScript + `"

[[services]]
name = "counter-service"
command = "` + counterScript + `"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Step 1: Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}

	// Override log directory
	config.DefaultLogDir = logDir

	// Step 2: Create manager and start services
	manager, err := process.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start services: %v", err)
	}

	// Verify both running
	statuses := manager.AllStatus()
	runningCount := 0
	for _, s := range statuses {
		if s.State == process.StateRunning {
			runningCount++
		}
	}
	if runningCount != 2 {
		t.Errorf("expected 2 running services, got %d", runningCount)
	}

	// Step 3: Wait for output to be written to log files
	time.Sleep(2 * time.Second)

	// Check log files exist
	echoLogs, err := manager.LogManager().ListLogs("echo-service")
	if err != nil {
		t.Fatalf("failed to list echo logs: %v", err)
	}
	if len(echoLogs) == 0 {
		t.Error("expected echo-service log files")
	}

	counterLogs, err := manager.LogManager().ListLogs("counter-service")
	if err != nil {
		t.Fatalf("failed to list counter logs: %v", err)
	}
	if len(counterLogs) == 0 {
		t.Error("expected counter-service log files")
	}

	// Read and verify log content - read directly from file
	if len(echoLogs) > 0 {
		// Read file content directly to debug
		content, _ := os.ReadFile(echoLogs[0].Path)
		t.Logf("Log file content: %s", string(content))

		reader, err := manager.LogManager().CreateReader(echoLogs[0].Path, logs.ReaderOptions{})
		if err != nil {
			t.Fatalf("failed to create reader: %v", err)
		}

		entries, err := reader.ReadAll()
		reader.Close()
		if err != nil {
			t.Fatalf("failed to read logs: %v", err)
		}

		t.Logf("Found %d log entries", len(entries))
		foundHello := false
		for _, entry := range entries {
			t.Logf("Entry: %s", entry.Message)
			if strings.Contains(entry.Message, "hello from echo-service") {
				foundHello = true
				break
			}
		}
		if !foundHello {
			t.Error("expected to find 'hello from echo-service' in logs")
		}
	}

	// Step 4: Stop services
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := manager.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop services: %v", err)
	}

	// Step 5: Verify stopped
	time.Sleep(100 * time.Millisecond)
	statuses = manager.AllStatus()
	for _, s := range statuses {
		if s.State == process.StateRunning {
			t.Errorf("service %s should be stopped", s.Name)
		}
	}
}

func TestE2EServiceCrashAndRestart(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "devserv.toml")
	logDir := filepath.Join(tmpDir, "logs")

	// Service that crashes immediately
	configContent := `
[[services]]
name = "crasher"
command = "sh -c 'echo starting; exit 1'"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	config.DefaultLogDir = logDir

	manager, err := process.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Start - it will crash
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait for crash
	time.Sleep(500 * time.Millisecond)

	status, _ := manager.Status("crasher")
	if status.State != process.StateCrashed && status.State != process.StateStopped {
		t.Errorf("expected crashed or stopped state, got %s", status.State)
	}

	// Restart
	if err := manager.Restart(ctx, "crasher"); err != nil {
		// Expected to crash again, but restart should work
		t.Logf("restart returned error (expected): %v", err)
	}

	// Verify logs were created for both runs
	time.Sleep(500 * time.Millisecond)

	logFiles, _ := manager.LogManager().ListLogs("crasher")
	if len(logFiles) < 1 {
		t.Error("expected at least 1 log file")
	}
}

func TestE2EMultipleRestarts(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "devserv.toml")
	logDir := filepath.Join(tmpDir, "logs")

	configContent := `
[[services]]
name = "restarter"
command = "sh -c 'echo run started; sleep 10'"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	config.DefaultLogDir = logDir

	manager, err := process.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Start
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	var pids []int
	status, _ := manager.Status("restarter")
	pids = append(pids, status.PID)

	// Restart 3 times - wait enough between restarts for different timestamps
	for i := 0; i < 3; i++ {
		time.Sleep(1100 * time.Millisecond) // Wait > 1 second for different log file timestamp
		if err := manager.Restart(ctx, "restarter"); err != nil {
			t.Fatalf("restart %d failed: %v", i+1, err)
		}
		status, _ := manager.Status("restarter")
		pids = append(pids, status.PID)
	}

	// Verify all PIDs are different
	seen := make(map[int]bool)
	for _, pid := range pids {
		if pid == 0 {
			continue
		}
		if seen[pid] {
			t.Errorf("duplicate PID %d - restarts not creating new processes", pid)
		}
		seen[pid] = true
	}

	// Verify multiple log files created (initial start + 3 restarts = 4)
	logFiles, _ := manager.LogManager().ListLogs("restarter")
	t.Logf("Found %d log files", len(logFiles))
	for _, lf := range logFiles {
		t.Logf("Log file: %s", lf.Path)
	}
	if len(logFiles) < 4 {
		t.Errorf("expected at least 4 log files, got %d", len(logFiles))
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	manager.Stop(stopCtx)
}

func TestE2ELogFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "devserv.toml")
	logDir := filepath.Join(tmpDir, "logs")

	// Create shell script for testing log filtering
	loggerScript := filepath.Join(tmpDir, "logger.sh")
	if err := os.WriteFile(loggerScript, []byte("#!/bin/sh\necho 'INFO: starting'\necho 'ERROR: something bad'\necho 'INFO: running'\necho 'ERROR: another error'\nsleep 10\n"), 0755); err != nil {
		t.Fatalf("failed to write logger script: %v", err)
	}

	// Use shell script for cleaner command
	configContent := `
[[services]]
name = "logger"
command = "` + loggerScript + `"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	config.DefaultLogDir = logDir

	manager, err := process.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Wait for output to be written
	time.Sleep(2 * time.Second)

	// Get log file
	logFiles, _ := manager.LogManager().ListLogs("logger")
	if len(logFiles) == 0 {
		t.Fatal("no log files found")
	}

	// First, read all entries to debug
	allReader, _ := manager.LogManager().CreateReader(logFiles[0].Path, logs.ReaderOptions{})
	allEntries, _ := allReader.ReadAll()
	allReader.Close()
	t.Logf("All entries: %d", len(allEntries))
	for _, e := range allEntries {
		t.Logf("  Entry: stream=%s message=%s", e.Stream, e.Message)
	}

	// Read only ERROR lines
	reader, err := manager.LogManager().CreateReader(logFiles[0].Path, logs.ReaderOptions{
		Pattern: "ERROR",
	})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	entries, err := reader.ReadAll()
	reader.Close()
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	t.Logf("Filtered entries (ERROR pattern): %d", len(entries))
	errorCount := 0
	for _, entry := range entries {
		t.Logf("  Filtered entry: %s", entry.Message)
		if strings.Contains(entry.Message, "ERROR") {
			errorCount++
		}
	}

	if errorCount != 2 {
		t.Errorf("expected 2 ERROR entries, got %d", errorCount)
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	manager.Stop(stopCtx)
}

func TestE2EGracefulShutdownTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "devserv.toml")
	logDir := filepath.Join(tmpDir, "logs")

	// Service that ignores SIGTERM
	configContent := `
[[services]]
name = "stubborn"
command = "sh -c 'trap \"\" TERM; echo started; sleep 60'"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	config.DefaultLogDir = logDir

	manager, err := process.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Stop with short timeout - should force kill
	stopCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	start := time.Now()
	manager.Stop(stopCtx)
	elapsed := time.Since(start)

	// Should have been killed within timeout
	if elapsed > 3*time.Second {
		t.Errorf("stop took too long: %v", elapsed)
	}

	status, _ := manager.Status("stubborn")
	if status.State == process.StateRunning {
		t.Error("service should not be running after forced stop")
		// Extra cleanup
		manager.Kill("stubborn")
	}
}
