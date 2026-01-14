package logs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReaderReadAll(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create log file with entries
	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	for i := 0; i < 10; i++ {
		writer.WriteLine("stdout", "message")
	}
	writer.Close()

	// Read all entries
	reader, err := NewReader(logPath, ReaderOptions{})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer reader.Close()

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read entries: %v", err)
	}

	if len(entries) != 10 {
		t.Errorf("expected 10 entries, got %d", len(entries))
	}
}

func TestReaderReadAllWithLinesLimit(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create log file with entries
	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	for i := 0; i < 100; i++ {
		writer.WriteLine("stdout", "message")
	}
	writer.Close()

	// Read with lines limit
	reader, err := NewReader(logPath, ReaderOptions{Lines: 10})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer reader.Close()

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read entries: %v", err)
	}

	if len(entries) != 10 {
		t.Errorf("expected 10 entries (last 10), got %d", len(entries))
	}
}

func TestReaderReadWithStreamFilter(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create log file with mixed streams
	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	for i := 0; i < 5; i++ {
		writer.WriteLine("stdout", "out message")
		writer.WriteLine("stderr", "err message")
	}
	writer.Close()

	// Read only stderr
	reader, err := NewReader(logPath, ReaderOptions{Stream: "stderr"})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer reader.Close()

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read entries: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("expected 5 stderr entries, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.Stream != "stderr" {
			t.Errorf("expected stream 'stderr', got '%s'", entry.Stream)
		}
	}
}

func TestReaderReadWithPatternFilter(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create log file
	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	writer.WriteLine("stdout", "INFO: starting server")
	writer.WriteLine("stdout", "ERROR: connection failed")
	writer.WriteLine("stdout", "INFO: server ready")
	writer.WriteLine("stdout", "ERROR: timeout")
	writer.Close()

	// Read only ERROR lines
	reader, err := NewReader(logPath, ReaderOptions{Pattern: "ERROR"})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer reader.Close()

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 ERROR entries, got %d", len(entries))
	}
}

func TestReaderReadChannel(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create log file
	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	for i := 0; i < 5; i++ {
		writer.WriteLine("stdout", "message")
	}
	writer.Close()

	// Read via channel
	reader, err := NewReader(logPath, ReaderOptions{})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count := 0
	for range reader.Read(ctx) {
		count++
	}

	if count != 5 {
		t.Errorf("expected 5 entries from channel, got %d", count)
	}
}

func TestReaderReadChannelWithCancel(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create log file with many entries
	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	for i := 0; i < 100; i++ {
		writer.WriteLine("stdout", "message")
	}
	writer.Close()

	// Read with early cancel
	reader, err := NewReader(logPath, ReaderOptions{})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	ch := reader.Read(ctx)
	count := 0

	for range ch {
		count++
		if count >= 10 {
			cancel()
			break
		}
	}

	// Should have stopped early
	if count > 20 {
		t.Errorf("expected early stop around 10 entries, got %d", count)
	}
}

func TestReaderInvalidPattern(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create empty log file
	os.WriteFile(logPath, []byte{}, 0644)

	// Invalid regex pattern
	_, err := NewReader(logPath, ReaderOptions{Pattern: "[invalid"})
	if err == nil {
		t.Error("expected error for invalid pattern")
	}
}
