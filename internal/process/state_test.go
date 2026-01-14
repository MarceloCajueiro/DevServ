package process

import "testing"

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateStopped, "stopped"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StateStopping, "stopping"},
		{StateCrashed, "crashed"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateSymbol(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateStopped, "○"},
		{StateStarting, "◐"},
		{StateRunning, "●"},
		{StateStopping, "◑"},
		{StateCrashed, "✖"},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.Symbol(); got != tt.expected {
				t.Errorf("State.Symbol() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateColor(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateStopped, "gray"},
		{StateStarting, "yellow"},
		{StateRunning, "green"},
		{StateStopping, "yellow"},
		{StateCrashed, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.Color(); got != tt.expected {
				t.Errorf("State.Color() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateIsActive(t *testing.T) {
	tests := []struct {
		state    State
		expected bool
	}{
		{StateStopped, false},
		{StateStarting, true},
		{StateRunning, true},
		{StateStopping, true},
		{StateCrashed, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.IsActive(); got != tt.expected {
				t.Errorf("State.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}
