// Package tui implements the interactive terminal user interface.
package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marcelocajueiro/devserv/internal/process"
)

// Run starts the TUI application.
func Run(manager *process.Manager) error {
	m := NewModel(manager)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Start background event listener with proper cleanup
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-manager.Events():
				if !ok {
					// Channel closed, exit gracefully
					return
				}
				p.Send(ServiceEventMsg{Event: event})
			}
		}
	}()

	_, err := p.Run()

	// Cancel context and wait for goroutine to finish
	cancel()
	<-done

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
