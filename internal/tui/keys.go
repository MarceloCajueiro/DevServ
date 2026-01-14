package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keyboard shortcuts.
type KeyMap struct {
	Quit      key.Binding
	Help      key.Binding
	Up        key.Binding
	Down      key.Binding
	Start     key.Binding
	Stop      key.Binding
	Restart   key.Binding
	StartAll  key.Binding
	StopAll   key.Binding
	Logs      key.Binding
	Back      key.Binding
	Kill      key.Binding
}

// Keys contains the default key bindings.
var Keys = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Start: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "start"),
	),
	Stop: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "stop"),
	),
	Restart: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "restart"),
	),
	StartAll: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "start all"),
	),
	StopAll: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "stop all"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l", "enter"),
		key.WithHelp("l", "logs"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Kill: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "force kill"),
	),
}

// ShortHelp returns key bindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Start, k.Stop, k.Restart, k.Logs, k.Help, k.Quit}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Start, k.Stop, k.Restart},
		{k.StartAll, k.StopAll, k.Kill},
		{k.Logs, k.Back},
		{k.Help, k.Quit},
	}
}
