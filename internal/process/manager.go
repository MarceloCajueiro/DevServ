package process

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/marcelocajueiro/devserv/internal/config"
	"github.com/marcelocajueiro/devserv/internal/logs"
	"github.com/marcelocajueiro/devserv/internal/state"
)

// Manager orchestrates multiple services.
type Manager struct {
	config        *config.Config
	services      map[string]*Service
	logManager    *logs.Manager
	eventsCh      chan Event
	stateEventsCh chan Event // Internal channel for state updates
	state         *state.State

	mu sync.RWMutex
}

// NewManager creates a new process manager.
func NewManager(cfg *config.Config) (*Manager, error) {
	logManager, err := logs.NewManager(config.DefaultLogDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create log manager: %w", err)
	}

	// Load shared state
	sharedState, err := state.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	stateEventsCh := make(chan Event, 100)

	m := &Manager{
		config:        cfg,
		services:      make(map[string]*Service),
		logManager:    logManager,
		eventsCh:      make(chan Event, 100),
		stateEventsCh: stateEventsCh,
		state:         sharedState,
	}

	// Initialize services and restore running state from shared state
	for _, svcCfg := range cfg.Services {
		svc := NewService(svcCfg, stateEventsCh)
		m.services[svcCfg.Name] = svc

		// Check if this service is already running (from another instance)
		if svcState := sharedState.Get(svcCfg.Name); svcState != nil {
			svc.RestoreFromState(svcState.PID, svcState.StartTime, svcState.LogFile)
		}
	}

	// Start background goroutine to update state on service events
	go m.watchEvents()

	return m, nil
}

// watchEvents listens for service events and updates shared state accordingly.
// It also forwards events to the public eventsCh for TUI consumption.
func (m *Manager) watchEvents() {
	for event := range m.stateEventsCh {
		// Update shared state when service stops or crashes
		switch event.Type {
		case EventStopped, EventCrashed:
			m.state.SetStopped(event.Service)
			m.state.Save() // Ignore error, best effort
		}

		// Forward event to public channel for TUI
		select {
		case m.eventsCh <- event:
		default:
			// Don't block if channel is full
		}
	}
}

// Start starts the specified services, or all services if none specified.
func (m *Manager) Start(ctx context.Context, names ...string) error {
	if len(names) == 0 {
		names = m.config.ServiceNames()
	}

	var errs []error
	for _, name := range names {
		if err := m.StartService(ctx, name); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to start some services: %v", errs)
	}
	return nil
}

// StartService starts a single service by name.
func (m *Manager) StartService(ctx context.Context, name string) error {
	m.mu.RLock()
	svc, ok := m.services[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("service not found: %s", name)
	}

	// Check port availability
	if svc.Config().Port > 0 {
		if err := checkPort(svc.Config().Port); err != nil {
			return fmt.Errorf("port %d is not available for service %s: %w", svc.Config().Port, name, err)
		}
	}

	if err := svc.Start(ctx, m.logManager); err != nil {
		return err
	}

	// Update shared state
	status := svc.Status()
	m.state.SetRunning(name, status.PID, status.Port, status.LogFile)
	if err := m.state.Save(); err != nil {
		// Log error but don't fail the start
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	return nil
}

// Stop stops the specified services, or all services if none specified.
func (m *Manager) Stop(ctx context.Context, names ...string) error {
	if len(names) == 0 {
		names = m.config.ServiceNames()
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(names))

	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			if err := m.StopService(ctx, n); err != nil {
				errCh <- err
			}
		}(name)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to stop some services: %v", errs)
	}
	return nil
}

// StopService stops a single service by name.
func (m *Manager) StopService(ctx context.Context, name string) error {
	m.mu.RLock()
	svc, ok := m.services[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("service not found: %s", name)
	}

	if err := svc.Stop(ctx); err != nil {
		return err
	}

	// Update shared state
	m.state.SetStopped(name)
	if err := m.state.Save(); err != nil {
		// Log error but don't fail the stop
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	return nil
}

// Restart restarts the specified services.
func (m *Manager) Restart(ctx context.Context, names ...string) error {
	if len(names) == 0 {
		names = m.config.ServiceNames()
	}

	for _, name := range names {
		if err := m.RestartService(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

// RestartService restarts a single service.
func (m *Manager) RestartService(ctx context.Context, name string) error {
	// Create a timeout context for stop
	stopCtx, cancel := context.WithTimeout(ctx, config.DefaultShutdownTimeout)
	defer cancel()

	if err := m.StopService(stopCtx, name); err != nil {
		return fmt.Errorf("failed to stop service %s: %w", name, err)
	}

	// Brief pause between stop and start
	time.Sleep(100 * time.Millisecond)

	return m.StartService(ctx, name)
}

// Kill forcefully terminates a service.
func (m *Manager) Kill(name string) error {
	m.mu.RLock()
	svc, ok := m.services[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("service not found: %s", name)
	}

	return svc.Kill()
}

// Status returns the status of a single service.
func (m *Manager) Status(name string) (Status, error) {
	m.mu.RLock()
	svc, ok := m.services[name]
	m.mu.RUnlock()

	if !ok {
		return Status{}, fmt.Errorf("service not found: %s", name)
	}

	return svc.Status(), nil
}

// AllStatus returns the status of all services.
func (m *Manager) AllStatus() []Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]Status, 0, len(m.services))
	for _, name := range m.config.ServiceNames() {
		if svc, ok := m.services[name]; ok {
			statuses = append(statuses, svc.Status())
		}
	}
	return statuses
}

// Events returns the events channel for subscribing to service events.
func (m *Manager) Events() <-chan Event {
	return m.eventsCh
}

// Shutdown stops all services and cleans up resources.
func (m *Manager) Shutdown(ctx context.Context) error {
	return m.Stop(ctx)
}

// ServiceNames returns the names of all configured services.
func (m *Manager) ServiceNames() []string {
	return m.config.ServiceNames()
}

// GetService returns a service by name.
func (m *Manager) GetService(name string) (*Service, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	svc, ok := m.services[name]
	return svc, ok
}

// LogManager returns the log manager.
func (m *Manager) LogManager() *logs.Manager {
	return m.logManager
}

// StateUpdatedAt returns when the shared state was last updated.
func (m *Manager) StateUpdatedAt() time.Time {
	return m.state.UpdatedAt
}

// RefreshState reloads the shared state from file.
// This is useful when another instance might have made changes.
func (m *Manager) RefreshState() error {
	newState, err := state.Load()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = newState

	// Update service states based on loaded state
	for name, svc := range m.services {
		if svcState := newState.Get(name); svcState != nil {
			// Service is running according to shared state
			status := svc.Status()
			if status.State != StateRunning {
				svc.RestoreFromState(svcState.PID, svcState.StartTime, svcState.LogFile)
			}
		} else {
			// Service is not in shared state, might have been stopped externally
			// We don't force stop here, just mark as stopped if PID is dead
		}
	}

	return nil
}

// checkPort checks if a port is available.
func checkPort(port int) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use", port)
	}
	listener.Close()
	return nil
}
