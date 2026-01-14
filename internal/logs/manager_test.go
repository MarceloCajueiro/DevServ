package logs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerCreateWriter(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	writer, err := manager.CreateWriter("test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	defer writer.Close()

	// Verify path structure
	path := writer.Path()
	if path == "" {
		t.Error("expected non-empty path")
	}

	// Should contain service name and date
	if !containsAll(path, "test-service", time.Now().Format("2006-01-02")) {
		t.Errorf("path doesn't match expected structure: %s", path)
	}

	// File should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("log file was not created")
	}
}

func TestManagerListLogs(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create multiple log files
	for i := 0; i < 3; i++ {
		writer, err := manager.CreateWriter("multi-log-service")
		if err != nil {
			t.Fatalf("failed to create writer %d: %v", i, err)
		}
		writer.WriteLine("stdout", "test message")
		writer.Close()
		time.Sleep(1100 * time.Millisecond) // Ensure different timestamps
	}

	// List logs
	logs, err := manager.ListLogs("multi-log-service")
	if err != nil {
		t.Fatalf("failed to list logs: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("expected 3 log files, got %d", len(logs))
	}

	// Should be sorted by time (newest first)
	for i := 1; i < len(logs); i++ {
		if logs[i].StartTime.After(logs[i-1].StartTime) {
			t.Error("logs not sorted by time (newest first)")
		}
	}
}

func TestManagerListLogsEmpty(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// List logs for nonexistent service
	logs, err := manager.ListLogs("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if logs != nil && len(logs) != 0 {
		t.Errorf("expected empty logs, got %d", len(logs))
	}
}

func TestManagerGetLatestLog(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create log files
	var lastPath string
	for i := 0; i < 3; i++ {
		writer, err := manager.CreateWriter("latest-service")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}
		lastPath = writer.Path()
		writer.Close()
		time.Sleep(1100 * time.Millisecond)
	}

	// Get latest
	latest, err := manager.GetLatestLog("latest-service")
	if err != nil {
		t.Fatalf("failed to get latest log: %v", err)
	}

	if latest == nil {
		t.Fatal("expected latest log file")
	}

	if latest.Path != lastPath {
		t.Errorf("expected path %s, got %s", lastPath, latest.Path)
	}
}

func TestManagerGetLatestLogNone(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	latest, err := manager.GetLatestLog("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if latest != nil {
		t.Error("expected nil for nonexistent service")
	}
}

func TestManagerCleanOldLogs(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create a log file
	writer, err := manager.CreateWriter("cleanup-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	logPath := writer.Path()
	writer.Close()

	// Verify it exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("log file should exist")
	}

	// Clean with 0 retention (delete all)
	if err := manager.CleanOldLogs(0); err != nil {
		t.Fatalf("failed to clean logs: %v", err)
	}

	// File should be deleted
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("log file should have been deleted")
	}
}

func TestManagerBaseDir(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	if manager.BaseDir() != dir {
		t.Errorf("expected base dir %s, got %s", dir, manager.BaseDir())
	}
}

func TestManagerExpandHomePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	// Use ~ prefix
	manager, err := NewManager("~/devserv-test-logs")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	expected := filepath.Join(home, "devserv-test-logs")
	if manager.BaseDir() != expected {
		t.Errorf("expected base dir %s, got %s", expected, manager.BaseDir())
	}

	// Cleanup
	os.RemoveAll(expected)
}

func TestManagerCreateReader(t *testing.T) {
	dir := t.TempDir()

	manager, err := NewManager(dir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create log file
	writer, err := manager.CreateWriter("reader-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	writer.WriteLine("stdout", "test message")
	logPath := writer.Path()
	writer.Close()

	// Create reader through manager
	reader, err := manager.CreateReader(logPath, ReaderOptions{})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer reader.Close()

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

// Helper function
func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
