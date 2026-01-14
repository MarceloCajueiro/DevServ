package logs

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriterWriteLine(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	// Write some lines
	if err := writer.WriteLine("stdout", "hello world"); err != nil {
		t.Fatalf("failed to write line: %v", err)
	}
	if err := writer.WriteLine("stderr", "error message"); err != nil {
		t.Fatalf("failed to write line: %v", err)
	}

	writer.Close()

	// Read and verify
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var entries []Entry

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse log entry: %v", err)
		}
		entries = append(entries, entry)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Check first entry
	if entries[0].Service != "test-service" {
		t.Errorf("expected service 'test-service', got '%s'", entries[0].Service)
	}
	if entries[0].Stream != "stdout" {
		t.Errorf("expected stream 'stdout', got '%s'", entries[0].Stream)
	}
	if entries[0].Message != "hello world" {
		t.Errorf("expected message 'hello world', got '%s'", entries[0].Message)
	}
	if entries[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	// Check second entry
	if entries[1].Stream != "stderr" {
		t.Errorf("expected stream 'stderr', got '%s'", entries[1].Stream)
	}
	if entries[1].Message != "error message" {
		t.Errorf("expected message 'error message', got '%s'", entries[1].Message)
	}
}

func TestWriterWriteEntry(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	now := time.Now()
	entry := Entry{
		Timestamp: now,
		Service:   "custom-service",
		Stream:    "stdout",
		Message:   "custom message",
	}

	if err := writer.WriteEntry(entry); err != nil {
		t.Fatalf("failed to write entry: %v", err)
	}

	writer.Close()

	// Read and verify
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	var readEntry Entry
	if err := json.Unmarshal(scanner.Bytes(), &readEntry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if readEntry.Service != "custom-service" {
		t.Errorf("expected service 'custom-service', got '%s'", readEntry.Service)
	}
	if readEntry.Message != "custom message" {
		t.Errorf("expected message 'custom message', got '%s'", readEntry.Message)
	}
}

func TestWriterPath(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	defer writer.Close()

	if writer.Path() != logPath {
		t.Errorf("expected path '%s', got '%s'", logPath, writer.Path())
	}
}

func TestWriterConcurrent(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	writer, err := NewWriter(logPath, "test-service")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	// Write concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				writer.WriteLine("stdout", "message")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	writer.Close()

	// Count lines
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}

	if count != 1000 {
		t.Errorf("expected 1000 lines, got %d", count)
	}
}
