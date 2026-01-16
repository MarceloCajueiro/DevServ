package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	ColorGreen  = lipgloss.Color("#00FF00")
	ColorYellow = lipgloss.Color("#FFFF00")
	ColorRed    = lipgloss.Color("#FF0000")
	ColorGray   = lipgloss.Color("#808080")
	ColorBlue   = lipgloss.Color("#0080FF")
	ColorWhite  = lipgloss.Color("#FFFFFF")
	ColorBlack  = lipgloss.Color("#000000")
	ColorDim    = lipgloss.Color("#666666")
)

// Styles
var (
	// Title style for the header
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorBlue).
			Padding(0, 1)

	// Subtle title (no background)
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGray)

	// Status bar at the bottom
	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(ColorWhite).
			Padding(0, 1)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Selected row
	SelectedStyle = lipgloss.NewStyle().
			Background(ColorBlue).
			Foreground(ColorWhite)

	// Normal row
	NormalStyle = lipgloss.NewStyle()

	// Status colors
	StatusRunning = lipgloss.NewStyle().Foreground(ColorGreen)
	StatusStarting = lipgloss.NewStyle().Foreground(ColorYellow)
	StatusStopped = lipgloss.NewStyle().Foreground(ColorGray)
	StatusCrashed = lipgloss.NewStyle().Foreground(ColorRed)

	// Log styles
	LogTimestamp = lipgloss.NewStyle().Foreground(ColorDim)
	LogStdout    = lipgloss.NewStyle().Foreground(ColorWhite)
	LogStderr    = lipgloss.NewStyle().Foreground(ColorRed)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)
)

// GetStatusStyle returns the appropriate style for a service state.
func GetStatusStyle(state string) lipgloss.Style {
	switch state {
	case "running":
		return StatusRunning
	case "starting", "stopping", "loading":
		return StatusStarting
	case "crashed":
		return StatusCrashed
	default:
		return StatusStopped
	}
}
