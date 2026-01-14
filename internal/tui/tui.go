// Package tui implements the interactive terminal user interface.
package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MarceloCajueiro/DevServ/internal/process"
)

// Run starts the TUI application.
func Run(manager *process.Manager) error {
	m := NewModel(manager)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Start background event listener
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-manager.Events():
				p.Send(ServiceEventMsg{Event: event})
			}
		}
	}()

	_, err := p.Run()
	return err
}

// Messages

// TickMsg is sent periodically to refresh the UI.
type TickMsg time.Time

// ServiceEventMsg wraps a process event for the TUI.
type ServiceEventMsg struct {
	Event process.Event
}

// ErrorMsg represents an error to display.
type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string {
	return e.Err.Error()
}
