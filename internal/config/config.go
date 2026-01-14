// Package config handles parsing and validation of devserv configuration files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the root configuration structure.
type Config struct {
	Services []Service `toml:"services"`
}

// Service represents a single service definition.
type Service struct {
	Name      string `toml:"name"`
	Command   string `toml:"command"`
	Directory string `toml:"directory,omitempty"`
	Port      int    `toml:"port,omitempty"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Services: []Service{},
	}
}

// Load reads and parses a configuration file from the given path.
// If path is empty, it searches in default locations.
func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = findConfigFile()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// findConfigFile searches for a config file in standard locations.
func findConfigFile() (string, error) {
	searchPaths := []string{
		"devserv.toml",
		"./devserv.toml",
	}

	// Add user config directories
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(home, ".config", "devserv", "config.toml"),
			filepath.Join(home, ".devserv", "config.toml"),
		)
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("no config file found (searched: devserv.toml, ~/.config/devserv/config.toml)")
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("at least one service must be defined")
	}

	seen := make(map[string]bool)
	for i, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("service %d: name is required", i+1)
		}
		if svc.Command == "" {
			return fmt.Errorf("service %q: command is required", svc.Name)
		}
		if seen[svc.Name] {
			return fmt.Errorf("duplicate service name: %q", svc.Name)
		}
		seen[svc.Name] = true
	}

	return nil
}

// GetService returns a service by name, or nil if not found.
func (c *Config) GetService(name string) *Service {
	for i := range c.Services {
		if c.Services[i].Name == name {
			return &c.Services[i]
		}
	}
	return nil
}

// ServiceNames returns a list of all service names.
func (c *Config) ServiceNames() []string {
	names := make([]string, len(c.Services))
	for i, svc := range c.Services {
		names[i] = svc.Name
	}
	return names
}

// Global settings with defaults
var (
	DefaultLogDir         = filepath.Join(os.Getenv("HOME"), ".devserv", "logs")
	DefaultPidDir         = filepath.Join(os.Getenv("HOME"), ".devserv", "pids")
	DefaultShutdownTimeout = 30 * time.Second
)
