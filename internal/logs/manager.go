// Package logs handles log file management, rotation, and reading.
package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Manager handles log file creation and organization.
type Manager struct {
	baseDir string
}

// LogFile represents metadata about a log file.
type LogFile struct {
	Path      string
	Service   string
	Date      string
	StartTime time.Time
	Size      int64
}

// NewManager creates a new log manager.
func NewManager(baseDir string) (*Manager, error) {
	// Expand ~ in path
	if strings.HasPrefix(baseDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, baseDir[1:])
	}

	// Create base directory if needed
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Manager{baseDir: baseDir}, nil
}

// CreateWriter creates a new log writer for a service.
// Creates directory structure: {baseDir}/{service}/{YYYY-MM-DD}/{HH-MM-SS}.log
func (m *Manager) CreateWriter(serviceName string) (*Writer, error) {
	now := time.Now()

	// Create directory path
	dir := filepath.Join(
		m.baseDir,
		serviceName,
		now.Format("2006-01-02"),
	)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create file with timestamp
	filename := now.Format("15-04-05") + ".log"
	path := filepath.Join(dir, filename)

	return NewWriter(path, serviceName)
}

// ListLogs returns all log files for a service.
func (m *Manager) ListLogs(serviceName string) ([]LogFile, error) {
	serviceDir := filepath.Join(m.baseDir, serviceName)

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return nil, nil // No logs yet
	}

	var logs []LogFile

	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".log") {
			return nil
		}

		// Parse date and time from path
		rel, _ := filepath.Rel(serviceDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))

		var date string
		var startTime time.Time

		if len(parts) >= 2 {
			date = parts[0]
			timeStr := strings.TrimSuffix(parts[1], ".log")
			if t, err := time.Parse("2006-01-02 15-04-05", date+" "+timeStr); err == nil {
				startTime = t
			}
		}

		logs = append(logs, LogFile{
			Path:      path,
			Service:   serviceName,
			Date:      date,
			StartTime: startTime,
			Size:      info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}

	// Sort by start time (most recent first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].StartTime.After(logs[j].StartTime)
	})

	return logs, nil
}

// GetLatestLog returns the most recent log file for a service.
func (m *Manager) GetLatestLog(serviceName string) (*LogFile, error) {
	logs, err := m.ListLogs(serviceName)
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return nil, nil
	}

	return &logs[0], nil
}

// CreateReader creates a reader for a specific log file.
func (m *Manager) CreateReader(path string, opts ReaderOptions) (*Reader, error) {
	return NewReader(path, opts)
}

// CleanOldLogs removes logs older than the specified duration.
func (m *Manager) CleanOldLogs(retention time.Duration) error {
	cutoff := time.Now().Add(-retention)

	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		serviceName := entry.Name()
		logs, err := m.ListLogs(serviceName)
		if err != nil {
			continue
		}

		for _, log := range logs {
			if log.StartTime.Before(cutoff) {
				os.Remove(log.Path)
			}
		}

		// Clean up empty date directories
		serviceDir := filepath.Join(m.baseDir, serviceName)
		dateDirs, _ := os.ReadDir(serviceDir)
		for _, dateDir := range dateDirs {
			if dateDir.IsDir() {
				datePath := filepath.Join(serviceDir, dateDir.Name())
				files, _ := os.ReadDir(datePath)
				if len(files) == 0 {
					os.Remove(datePath)
				}
			}
		}
	}

	return nil
}

// BaseDir returns the base log directory.
func (m *Manager) BaseDir() string {
	return m.baseDir
}
