package process

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/marcelocajueiro/devserv/internal/config"
	"github.com/marcelocajueiro/devserv/internal/logs"
)

// Integration tests with real processes

func TestServiceStartStop(t *testing.T) {
	// Create temp log directory
	logDir := t.TempDir()
	logManager, err := logs.NewManager(logDir)
	if err != nil {
		t.Fatalf("failed to create log manager: %v", err)
	}

	// Create service with a simple sleep command
	cfg := config.Service{
		Name:    "test-sleep",
		Command: "sleep 60",
	}

	eventsCh := make(chan Event, 10)
	svc := NewService(cfg, eventsCh)

	// Start
	ctx := context.Background()
	if err := svc.Start(ctx, logManager); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Verify running
	status := svc.Status()
	if status.State != StateRunning {
		t.Errorf("expected state Running, got %s", status.State)
	}
	if status.PID <= 0 {
		t.Error("expected valid PID")
	}

	// Check event
	select {
	case event := <-eventsCh:
		if event.Type != EventStarted {
			t.Errorf("expected EventStarted, got %v", event.Type)
		}
		if event.Service != "test-sleep" {
			t.Errorf("expected service 'test-sleep', got '%s'", event.Service)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for start event")
	}

	// Stop
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := svc.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop service: %v", err)
	}

	// Wait a bit for state to update
	time.Sleep(100 * time.Millisecond)

	// Verify stopped
	status = svc.Status()
	if status.State != StateStopped {
		t.Errorf("expected state Stopped, got %s", status.State)
	}
}

func TestServiceCapturesOutput(t *testing.T) {
	logDir := t.TempDir()
	logManager, err := logs.NewManager(logDir)
	if err != nil {
		t.Fatalf("failed to create log manager: %v", err)
	}

	// Create a shell script for testing output capture
	scriptPath := filepath.Join(t.TempDir(), "output.sh")
	scriptContent := "#!/bin/sh\necho stdout-message\necho stderr-message >&2\nsleep 5\n"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	// Service that outputs to stdout and stderr
	cfg := config.Service{
		Name:    "test-echo",
		Command: scriptPath,
	}

	eventsCh := make(chan Event, 100)
	svc := NewService(cfg, eventsCh)

	ctx := context.Background()
	if err := svc.Start(ctx, logManager); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait for output events
	var stdoutFound, stderrFound bool
	timeout := time.After(5 * time.Second)

loop:
	for {
		select {
		case event := <-eventsCh:
			if event.Type == EventOutput {
				data := event.Data.(map[string]string)
				if data["stream"] == "stdout" && strings.Contains(data["line"], "stdout-message") {
					stdoutFound = true
				}
				if data["stream"] == "stderr" && strings.Contains(data["line"], "stderr-message") {
					stderrFound = true
				}
				if stdoutFound && stderrFound {
					break loop
				}
			}
		case <-timeout:
			break loop
		}
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	svc.Stop(stopCtx)

	if !stdoutFound {
		t.Error("stdout message not captured")
	}
	if !stderrFound {
		t.Error("stderr message not captured")
	}

	// Verify log file exists
	status := svc.Status()
	if status.LogFile == "" {
		t.Error("expected log file path")
	}
}

func TestServiceCrashDetection(t *testing.T) {
	logDir := t.TempDir()
	logManager, err := logs.NewManager(logDir)
	if err != nil {
		t.Fatalf("failed to create log manager: %v", err)
	}

	// Service that exits immediately with error
	cfg := config.Service{
		Name:    "test-crash",
		Command: "sh -c 'exit 1'",
	}

	eventsCh := make(chan Event, 10)
	svc := NewService(cfg, eventsCh)

	ctx := context.Background()
	if err := svc.Start(ctx, logManager); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait for crash event
	var crashEvent *Event
	timeout := time.After(5 * time.Second)

loop:
	for {
		select {
		case event := <-eventsCh:
			if event.Type == EventCrashed {
				crashEvent = &event
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	if crashEvent == nil {
		t.Fatal("expected crash event")
	}

	// Wait for state to update
	time.Sleep(100 * time.Millisecond)

	status := svc.Status()
	if status.State != StateCrashed {
		t.Errorf("expected state Crashed, got %s", status.State)
	}
}

func TestServiceWorkingDirectory(t *testing.T) {
	logDir := t.TempDir()
	logManager, err := logs.NewManager(logDir)
	if err != nil {
		t.Fatalf("failed to create log manager: %v", err)
	}

	// Create a temp directory and a file in it
	workDir := t.TempDir()
	testFile := filepath.Join(workDir, "testfile.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Service that lists files in working directory
	cfg := config.Service{
		Name:      "test-pwd",
		Command:   "sh -c 'ls testfile.txt && pwd'",
		Directory: workDir,
	}

	eventsCh := make(chan Event, 100)
	svc := NewService(cfg, eventsCh)

	ctx := context.Background()
	if err := svc.Start(ctx, logManager); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait for output
	var foundFile bool
	timeout := time.After(5 * time.Second)

loop:
	for {
		select {
		case event := <-eventsCh:
			if event.Type == EventOutput {
				data := event.Data.(map[string]string)
				if strings.Contains(data["line"], "testfile.txt") {
					foundFile = true
					break loop
				}
			}
			if event.Type == EventStopped || event.Type == EventCrashed {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	// Cleanup
	time.Sleep(100 * time.Millisecond)

	if !foundFile {
		t.Error("working directory not set correctly - testfile.txt not found")
	}
}

func TestServiceKill(t *testing.T) {
	logDir := t.TempDir()
	logManager, err := logs.NewManager(logDir)
	if err != nil {
		t.Fatalf("failed to create log manager: %v", err)
	}

	// Service that ignores SIGTERM (traps it)
	cfg := config.Service{
		Name:    "test-stubborn",
		Command: "sh -c 'trap \"\" TERM; sleep 60'",
	}

	eventsCh := make(chan Event, 10)
	svc := NewService(cfg, eventsCh)

	ctx := context.Background()
	if err := svc.Start(ctx, logManager); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait for it to start
	time.Sleep(100 * time.Millisecond)

	// Force kill
	if err := svc.Kill(); err != nil {
		t.Fatalf("failed to kill service: %v", err)
	}

	// Wait for it to die
	time.Sleep(200 * time.Millisecond)

	status := svc.Status()
	if status.State == StateRunning {
		t.Error("service should not be running after kill")
	}
}

func TestServiceAlreadyRunning(t *testing.T) {
	logDir := t.TempDir()
	logManager, err := logs.NewManager(logDir)
	if err != nil {
		t.Fatalf("failed to create log manager: %v", err)
	}

	cfg := config.Service{
		Name:    "test-double-start",
		Command: "sleep 60",
	}

	svc := NewService(cfg, nil)

	ctx := context.Background()
	if err := svc.Start(ctx, logManager); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Try to start again - should be a no-op (no error)
	err = svc.Start(ctx, logManager)
	if err != nil {
		t.Errorf("expected no error when starting already running service, got: %v", err)
	}

	// Verify service is still running
	status := svc.Status()
	if status.State != StateRunning {
		t.Errorf("expected state Running, got %v", status.State)
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	svc.Stop(stopCtx)
}

func TestServiceUptime(t *testing.T) {
	logDir := t.TempDir()
	logManager, err := logs.NewManager(logDir)
	if err != nil {
		t.Fatalf("failed to create log manager: %v", err)
	}

	cfg := config.Service{
		Name:    "test-uptime",
		Command: "sleep 60",
	}

	svc := NewService(cfg, nil)

	ctx := context.Background()
	if err := svc.Start(ctx, logManager); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	status := svc.Status()
	if status.Uptime < 400*time.Millisecond {
		t.Errorf("expected uptime > 400ms, got %v", status.Uptime)
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	svc.Stop(stopCtx)
}
