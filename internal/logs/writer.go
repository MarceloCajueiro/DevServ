package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Writer writes structured log entries to a file.
type Writer struct {
	file    *os.File
	encoder *json.Encoder
	path    string
	service string
	mu      sync.Mutex
}

// Entry represents a single log entry.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Stream    string    `json:"stream"` // "stdout" or "stderr"
	Message   string    `json:"message"`
}

// NewWriter creates a new log writer.
func NewWriter(path, service string) (*Writer, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Writer{
		file:    file,
		encoder: json.NewEncoder(file),
		path:    path,
		service: service,
	}, nil
}

// WriteLine writes a single line to the log.
func (w *Writer) WriteLine(stream, message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry := Entry{
		Timestamp: time.Now(),
		Service:   w.service,
		Stream:    stream,
		Message:   message,
	}

	return w.encoder.Encode(entry)
}

// WriteEntry writes a log entry directly.
func (w *Writer) WriteEntry(entry Entry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.Service == "" {
		entry.Service = w.service
	}

	return w.encoder.Encode(entry)
}

// Path returns the log file path.
func (w *Writer) Path() string {
	return w.path
}

// Close closes the log writer.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Sync flushes the log file to disk.
func (w *Writer) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}
