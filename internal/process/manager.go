package process

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/MarceloCajueiro/DevServ/internal/config"
	"github.com/MarceloCajueiro/DevServ/internal/logs"
)

// Manager orchestrates multiple services.
type Manager struct {
	config     *config.Config
	services   map[string]*Service
	logManager *logs.Manager
	eventsCh   chan Event

	mu sync.RWMutex
}

// NewManager creates a new process manager.
func NewManager(cfg *config.Config) (*Manager, error) {
	logManager, err := logs.NewManager(config.DefaultLogDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create log manager: %w", err)
	}

	m := &Manager{
		config:     cfg,
		services:   make(map[string]*Service),
		logManager: logManager,
		eventsCh:   make(chan Event, 100),
	}

	// Initialize services
	for _, svcCfg := range cfg.Services {
		m.services[svcCfg.Name] = NewService(svcCfg, m.eventsCh)
	}

	return m, nil
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

	return svc.Start(ctx, m.logManager)
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

	return svc.Stop(ctx)
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
