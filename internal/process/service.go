package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/marcelocajueiro/devserv/internal/config"
	"github.com/marcelocajueiro/devserv/internal/logs"
)

// Service represents a managed process.
type Service struct {
	config    config.Service
	state     State
	cmd       *exec.Cmd
	pid       int
	startTime time.Time
	exitCode  int
	error     string
	logWriter *logs.Writer

	mu       sync.RWMutex
	done     chan struct{}
	eventsCh chan<- Event
}

// Status contains the current status of a service.
type Status struct {
	Name      string
	State     State
	PID       int
	Port      int
	StartTime time.Time
	Uptime    time.Duration
	ExitCode  int
	Error     string
	LogFile   string
}

// Event represents a service lifecycle event.
type Event struct {
	Service   string
	Type      EventType
	Timestamp time.Time
	Data      interface{}
}

// EventType represents the type of service event.
type EventType int

const (
	EventStarted EventType = iota
	EventStopped
	EventCrashed
	EventOutput
)

// NewService creates a new service instance.
func NewService(cfg config.Service, eventsCh chan<- Event) *Service {
	return &Service{
		config:   cfg,
		state:    StateUnknown,
		eventsCh: eventsCh,
	}
}

// RestoreFromState restores a service state from shared state file.
// This is used when another instance has started the service.
func (s *Service) RestoreFromState(pid int, startTime time.Time, logFile string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pid = pid
	s.startTime = startTime
	s.state = StateRunning

	// Note: We can't restore cmd or logWriter for external processes,
	// but we can track that it's running via PID
}

// MarkAsStopped marks the service as stopped.
// This is used when another instance has stopped the service.
func (s *Service) MarkAsStopped() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only update if we don't own the process
	if s.cmd == nil {
		s.state = StateStopped
		s.pid = 0
		s.error = ""
	}
}

// MarkAsCrashed marks the service as crashed.
// This is used when another instance detected a crash.
func (s *Service) MarkAsCrashed(errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only update if we don't own the process
	if s.cmd == nil {
		s.state = StateCrashed
		s.pid = 0
		s.error = errMsg
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return s.config.Name
}

// Config returns the service configuration.
func (s *Service) Config() config.Service {
	return s.config
}

// Status returns the current service status.
func (s *Service) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := Status{
		Name:      s.config.Name,
		State:     s.state,
		PID:       s.pid,
		Port:      s.config.Port,
		StartTime: s.startTime,
		ExitCode:  s.exitCode,
		Error:     s.error,
	}

	if s.state == StateRunning && !s.startTime.IsZero() {
		status.Uptime = time.Since(s.startTime)
	}

	if s.logWriter != nil {
		status.LogFile = s.logWriter.Path()
	}

	return status
}

// Start starts the service.
func (s *Service) Start(ctx context.Context, logManager *logs.Manager) error {
	s.mu.Lock()
	if s.state == StateRunning || s.state == StateStarting {
		s.mu.Unlock()
		return fmt.Errorf("service %s is already running", s.config.Name)
	}

	s.state = StateStarting
	s.error = ""
	s.exitCode = 0
	s.mu.Unlock()

	// Create log writer for this session
	logWriter, err := logManager.CreateWriter(s.config.Name)
	if err != nil {
		s.mu.Lock()
		s.state = StateStopped
		s.error = err.Error()
		s.mu.Unlock()
		return fmt.Errorf("failed to create log writer: %w", err)
	}

	// Parse command
	parts := strings.Fields(s.config.Command)
	if len(parts) == 0 {
		logWriter.Close()
		s.mu.Lock()
		s.state = StateStopped
		s.error = "empty command"
		s.mu.Unlock()
		return fmt.Errorf("empty command for service %s", s.config.Name)
	}

	// Create command
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	// Set working directory
	if s.config.Directory != "" {
		dir := s.config.Directory
		if !filepath.IsAbs(dir) {
			if cwd, err := os.Getwd(); err == nil {
				dir = filepath.Join(cwd, dir)
			}
		}
		cmd.Dir = dir
	}

	// Create a new process group so we can kill all child processes
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Set up pipes for stdout/stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logWriter.Close()
		s.mu.Lock()
		s.state = StateStopped
		s.error = err.Error()
		s.mu.Unlock()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logWriter.Close()
		s.mu.Lock()
		s.state = StateStopped
		s.error = err.Error()
		s.mu.Unlock()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		logWriter.Close()
		s.mu.Lock()
		s.state = StateStopped
		s.error = err.Error()
		s.mu.Unlock()
		return fmt.Errorf("failed to start service %s: %w", s.config.Name, err)
	}

	s.mu.Lock()
	s.cmd = cmd
	s.pid = cmd.Process.Pid
	s.startTime = time.Now()
	s.state = StateRunning
	s.logWriter = logWriter
	s.done = make(chan struct{})
	s.mu.Unlock()

	// Send started event
	s.sendEvent(EventStarted, nil)

	// Start goroutines to capture output
	go s.captureOutput(stdout, "stdout", logWriter)
	go s.captureOutput(stderr, "stderr", logWriter)

	// Start goroutine to wait for process exit
	go s.waitForExit()

	return nil
}

// Stop stops the service gracefully.
func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.state != StateRunning {
		s.mu.Unlock()
		return nil
	}

	s.state = StateStopping
	cmd := s.cmd
	done := s.done
	pid := s.pid
	s.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Send SIGTERM to the process group (negative PID)
	// This kills all child processes as well
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		// Process might have already exited
		if err != syscall.ESRCH {
			return fmt.Errorf("failed to send SIGTERM: %w", err)
		}
	}

	// Wait for process to exit or timeout
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Force kill the process group
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
			if err != syscall.ESRCH {
				return fmt.Errorf("failed to send SIGKILL: %w", err)
			}
		}
		<-done
		return nil
	}
}

// Kill forcefully terminates the service.
func (s *Service) Kill() error {
	s.mu.RLock()
	cmd := s.cmd
	pid := s.pid
	s.mu.RUnlock()

	if cmd == nil || cmd.Process == nil || pid <= 0 {
		return nil
	}

	// Kill the entire process group
	return syscall.Kill(-pid, syscall.SIGKILL)
}

func (s *Service) captureOutput(r io.Reader, stream string, logWriter *logs.Writer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Write to log file
		logWriter.WriteLine(stream, line)

		// Send output event
		s.sendEvent(EventOutput, map[string]string{
			"stream": stream,
			"line":   line,
		})
	}
}

func (s *Service) waitForExit() {
	s.mu.RLock()
	cmd := s.cmd
	done := s.done
	s.mu.RUnlock()

	if cmd == nil {
		return
	}

	err := cmd.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Close done channel
	if done != nil {
		close(done)
	}

	// Close log writer
	if s.logWriter != nil {
		s.logWriter.Close()
	}

	// Update state
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.exitCode = exitErr.ExitCode()
		}
		// Check if it was a graceful stop
		if s.state == StateStopping {
			s.state = StateStopped
			s.sendEvent(EventStopped, nil)
		} else {
			s.state = StateCrashed
			s.error = err.Error()
			s.sendEvent(EventCrashed, err.Error())
		}
	} else {
		s.state = StateStopped
		s.exitCode = 0
		s.sendEvent(EventStopped, nil)
	}

	s.cmd = nil
	s.pid = 0
}

func (s *Service) sendEvent(eventType EventType, data interface{}) {
	if s.eventsCh == nil {
		return
	}

	select {
	case s.eventsCh <- Event{
		Service:   s.config.Name,
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}:
	default:
		// Don't block if channel is full
	}
}
