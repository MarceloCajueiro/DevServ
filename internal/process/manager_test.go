package process

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/marcelomd/devserv/internal/config"
)

func TestManagerStartStop(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "svc1", Command: "sleep 60"},
			{Name: "svc2", Command: "sleep 60"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Start all
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start services: %v", err)
	}

	// Check status
	statuses := manager.AllStatus()
	if len(statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(statuses))
	}

	for _, status := range statuses {
		if status.State != StateRunning {
			t.Errorf("expected service %s to be running, got %s", status.Name, status.State)
		}
	}

	// Stop all
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := manager.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop services: %v", err)
	}

	// Verify stopped
	time.Sleep(100 * time.Millisecond)
	statuses = manager.AllStatus()
	for _, status := range statuses {
		if status.State == StateRunning {
			t.Errorf("expected service %s to be stopped", status.Name)
		}
	}
}

func TestManagerStartSpecific(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "svc1", Command: "sleep 60"},
			{Name: "svc2", Command: "sleep 60"},
			{Name: "svc3", Command: "sleep 60"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Start only svc1 and svc3
	if err := manager.Start(ctx, "svc1", "svc3"); err != nil {
		t.Fatalf("failed to start services: %v", err)
	}

	// Check status
	status1, _ := manager.Status("svc1")
	status2, _ := manager.Status("svc2")
	status3, _ := manager.Status("svc3")

	if status1.State != StateRunning {
		t.Errorf("expected svc1 to be running, got %s", status1.State)
	}
	if status2.State != StateStopped {
		t.Errorf("expected svc2 to be stopped, got %s", status2.State)
	}
	if status3.State != StateRunning {
		t.Errorf("expected svc3 to be running, got %s", status3.State)
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	manager.Stop(stopCtx)
}

func TestManagerRestart(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "svc1", Command: "sleep 60"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Start
	if err := manager.Start(ctx, "svc1"); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	status1, _ := manager.Status("svc1")
	oldPID := status1.PID

	// Restart
	if err := manager.Restart(ctx, "svc1"); err != nil {
		t.Fatalf("failed to restart service: %v", err)
	}

	status2, _ := manager.Status("svc1")
	newPID := status2.PID

	if oldPID == newPID {
		t.Error("expected different PID after restart")
	}

	if status2.State != StateRunning {
		t.Errorf("expected service to be running after restart, got %s", status2.State)
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	manager.Stop(stopCtx)
}

func TestManagerPortConflict(t *testing.T) {
	// Occupy a port
	listener, err := net.Listen("tcp", ":18080")
	if err != nil {
		t.Fatalf("failed to occupy port: %v", err)
	}
	defer listener.Close()

	cfg := &config.Config{
		Services: []config.Service{
			{Name: "svc1", Command: "sleep 60", Port: 18080},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Start should fail due to port conflict
	err = manager.Start(ctx, "svc1")
	if err == nil {
		t.Error("expected error due to port conflict")
		// Cleanup if it somehow started
		stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		manager.Stop(stopCtx)
	}
}

func TestManagerServiceNotFound(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "svc1", Command: "sleep 60"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Try to start nonexistent service
	err = manager.Start(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent service")
	}

	// Try to stop nonexistent service
	err = manager.StopService(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent service")
	}

	// Try to get status of nonexistent service
	_, err = manager.Status("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent service")
	}
}

func TestManagerKill(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "stubborn", Command: "sh -c 'trap \"\" TERM; sleep 60'"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Start
	if err := manager.Start(ctx, "stubborn"); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Kill
	if err := manager.Kill("stubborn"); err != nil {
		t.Fatalf("failed to kill service: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	status, _ := manager.Status("stubborn")
	if status.State == StateRunning {
		t.Error("service should not be running after kill")
	}
}

func TestManagerServiceNames(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "api", Command: "sleep 1"},
			{Name: "worker", Command: "sleep 1"},
			{Name: "db", Command: "sleep 1"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	names := manager.ServiceNames()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}

	expected := map[string]bool{"api": true, "worker": true, "db": true}
	for _, name := range names {
		if !expected[name] {
			t.Errorf("unexpected service name: %s", name)
		}
	}
}

func TestManagerEvents(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "svc1", Command: "sleep 60"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	eventsCh := manager.Events()
	ctx := context.Background()

	// Start
	if err := manager.Start(ctx, "svc1"); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Should receive start event
	select {
	case event := <-eventsCh:
		if event.Type != EventStarted {
			t.Errorf("expected EventStarted, got %v", event.Type)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for event")
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	manager.Stop(stopCtx)
}

func TestManagerGetService(t *testing.T) {
	cfg := &config.Config{
		Services: []config.Service{
			{Name: "api", Command: "sleep 1"},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Found
	svc, ok := manager.GetService("api")
	if !ok {
		t.Error("expected to find service 'api'")
	}
	if svc.Name() != "api" {
		t.Errorf("expected name 'api', got '%s'", svc.Name())
	}

	// Not found
	_, ok = manager.GetService("nonexistent")
	if ok {
		t.Error("expected not to find 'nonexistent' service")
	}
}
