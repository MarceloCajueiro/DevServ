// Package state manages shared state between DevServ instances via a JSON file.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// statePathOverride allows tests to use a custom state file path.
var statePathOverride string

// SetStatePath sets a custom path for the state file (for testing).
func SetStatePath(path string) {
	statePathOverride = path
}

// StatePath returns the path to the state file.
func StatePath() string {
	if statePathOverride != "" {
		return statePathOverride
	}
	return filepath.Join(os.Getenv("HOME"), ".devserv", "state.json")
}

// ServiceState represents the state of a single service.
type ServiceState struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "running", "stopped", "crashed"
	PID       int       `json:"pid,omitempty"`
	Port      int       `json:"port,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	LogFile   string    `json:"log_file,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// State represents the shared state of all services.
type State struct {
	Services  map[string]*ServiceState `json:"services"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// Load reads the state file. Returns empty state if file doesn't exist.
func Load() (*State, error) {
	path := StatePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{
				Services:  make(map[string]*ServiceState),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// Corrupted file, start fresh
		return &State{
			Services:  make(map[string]*ServiceState),
			UpdatedAt: time.Now(),
		}, nil
	}

	if state.Services == nil {
		state.Services = make(map[string]*ServiceState)
	}

	return &state, nil
}

// Save writes the state to file with file locking.
func (s *State) Save() error {
	path := StatePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Open file with lock
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()

	// Acquire exclusive lock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to lock state file: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	// Update timestamp
	s.UpdatedAt = time.Now()

	// Write state
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Truncate and write
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate state file: %w", err)
	}
	if _, err := file.WriteAt(data, 0); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// SetRunning marks a service as running.
func (s *State) SetRunning(name string, pid, port int, logFile string) {
	s.Services[name] = &ServiceState{
		Name:      name,
		Status:    "running",
		PID:       pid,
		Port:      port,
		StartTime: time.Now(),
		LogFile:   logFile,
	}
}

// SetStopped marks a service as stopped.
func (s *State) SetStopped(name string) {
	s.Services[name] = &ServiceState{
		Name:   name,
		Status: "stopped",
	}
}

// SetCrashed marks a service as crashed.
func (s *State) SetCrashed(name string, err string) {
	s.Services[name] = &ServiceState{
		Name:   name,
		Status: "crashed",
		Error:  err,
	}
}

// Get returns the state of a service, or nil if not found.
func (s *State) Get(name string) *ServiceState {
	svc, ok := s.Services[name]
	if !ok {
		return nil
	}
	return svc
}

// IsRunning checks if a service is running.
func (s *State) IsRunning(name string) bool {
	svc := s.Get(name)
	return svc != nil && svc.Status == "running"
}

// GetPID returns the PID of a running service, or 0 if not running.
func (s *State) GetPID(name string) int {
	svc := s.Get(name)
	if svc == nil {
		return 0
	}
	return svc.PID
}

// AllRunning returns all currently running services.
func (s *State) AllRunning() []*ServiceState {
	result := make([]*ServiceState, 0, len(s.Services))
	for _, svc := range s.Services {
		if svc.Status == "running" {
			result = append(result, svc)
		}
	}
	return result
}

// isProcessAlive checks if a process with the given PID exists.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
