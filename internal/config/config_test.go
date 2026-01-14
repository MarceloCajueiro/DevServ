package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	// Create temp config file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "devserv.toml")

	content := `
[[services]]
name = "api"
command = "go run ./cmd/api"
directory = "./backend"
port = 8080

[[services]]
name = "worker"
command = "python worker.py"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cfg.Services))
	}

	// Check first service
	if cfg.Services[0].Name != "api" {
		t.Errorf("expected name 'api', got '%s'", cfg.Services[0].Name)
	}
	if cfg.Services[0].Command != "go run ./cmd/api" {
		t.Errorf("expected command 'go run ./cmd/api', got '%s'", cfg.Services[0].Command)
	}
	if cfg.Services[0].Directory != "./backend" {
		t.Errorf("expected directory './backend', got '%s'", cfg.Services[0].Directory)
	}
	if cfg.Services[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Services[0].Port)
	}

	// Check second service (no directory/port)
	if cfg.Services[1].Name != "worker" {
		t.Errorf("expected name 'worker', got '%s'", cfg.Services[1].Name)
	}
	if cfg.Services[1].Directory != "" {
		t.Errorf("expected empty directory, got '%s'", cfg.Services[1].Directory)
	}
	if cfg.Services[1].Port != 0 {
		t.Errorf("expected port 0, got %d", cfg.Services[1].Port)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/devserv.toml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "devserv.toml")

	content := `this is not valid toml [[[`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestValidateEmptyServices(t *testing.T) {
	cfg := &Config{
		Services: []Service{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for empty services")
	}
}

func TestValidateMissingName(t *testing.T) {
	cfg := &Config{
		Services: []Service{
			{Command: "echo hello"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestValidateMissingCommand(t *testing.T) {
	cfg := &Config{
		Services: []Service{
			{Name: "test"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestValidateDuplicateName(t *testing.T) {
	cfg := &Config{
		Services: []Service{
			{Name: "api", Command: "echo 1"},
			{Name: "api", Command: "echo 2"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := &Config{
		Services: []Service{
			{Name: "api", Command: "echo hello"},
			{Name: "worker", Command: "echo world"},
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetService(t *testing.T) {
	cfg := &Config{
		Services: []Service{
			{Name: "api", Command: "echo api"},
			{Name: "worker", Command: "echo worker"},
		},
	}

	// Found
	svc := cfg.GetService("api")
	if svc == nil {
		t.Error("expected to find 'api' service")
	}
	if svc.Command != "echo api" {
		t.Errorf("expected command 'echo api', got '%s'", svc.Command)
	}

	// Not found
	svc = cfg.GetService("nonexistent")
	if svc != nil {
		t.Error("expected nil for nonexistent service")
	}
}

func TestServiceNames(t *testing.T) {
	cfg := &Config{
		Services: []Service{
			{Name: "api", Command: "echo api"},
			{Name: "worker", Command: "echo worker"},
			{Name: "db", Command: "echo db"},
		},
	}

	names := cfg.ServiceNames()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}

	expected := []string{"api", "worker", "db"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected name '%s' at index %d, got '%s'", expected[i], i, name)
		}
	}
}
