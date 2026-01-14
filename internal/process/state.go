// Package process handles service lifecycle and process management.
package process

// State represents the current state of a service.
type State int

const (
	// StateStopped indicates the service is not running.
	StateStopped State = iota
	// StateStarting indicates the service is being started.
	StateStarting
	// StateRunning indicates the service is running.
	StateRunning
	// StateStopping indicates the service is being stopped.
	StateStopping
	// StateCrashed indicates the service exited unexpectedly.
	StateCrashed
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateCrashed:
		return "crashed"
	default:
		return "unknown"
	}
}

// Symbol returns a visual symbol for the state.
func (s State) Symbol() string {
	switch s {
	case StateStopped:
		return "○"
	case StateStarting:
		return "◐"
	case StateRunning:
		return "●"
	case StateStopping:
		return "◑"
	case StateCrashed:
		return "✖"
	default:
		return "?"
	}
}

// Color returns the ANSI color name for the state.
func (s State) Color() string {
	switch s {
	case StateRunning:
		return "green"
	case StateStarting, StateStopping:
		return "yellow"
	case StateCrashed:
		return "red"
	default:
		return "gray"
	}
}

// IsActive returns true if the service is running or transitioning.
func (s State) IsActive() bool {
	return s == StateRunning || s == StateStarting || s == StateStopping
}
