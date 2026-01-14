package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marcelocajueiro/devserv/internal/process"
)

// ViewMode represents the current view.
type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewLogs
	ViewHelp
)

// Model is the main TUI model.
type Model struct {
	manager   *process.Manager
	viewMode  ViewMode
	selected  int
	statuses  []process.Status
	width     int
	height    int
	help      help.Model
	showHelp  bool
	err       error
	message   string
	msgExpiry time.Time

	// Logs view state
	logsService string
	logEntries  []string
	logOffset   int
}

// NewModel creates a new TUI model.
func NewModel(manager *process.Manager) *Model {
	h := help.New()
	h.ShowAll = false

	return &Model{
		manager:  manager,
		viewMode: ViewDashboard,
		statuses: manager.AllStatus(),
		help:     h,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
	)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case TickMsg:
		m.statuses = m.manager.AllStatus()
		// Clear expired messages
		if !m.msgExpiry.IsZero() && time.Now().After(m.msgExpiry) {
			m.message = ""
			m.msgExpiry = time.Time{}
		}
		return m, m.tickCmd()

	case ServiceEventMsg:
		m.handleServiceEvent(msg.Event)
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle help toggle
	if key.Matches(msg, Keys.Help) {
		m.showHelp = !m.showHelp
		return m, nil
	}

	// If help is shown, only allow closing it
	if m.showHelp {
		if key.Matches(msg, Keys.Back, Keys.Quit) {
			m.showHelp = false
		}
		return m, nil
	}

	// Handle view-specific keys
	switch m.viewMode {
	case ViewLogs:
		return m.handleLogsKey(msg)
	default:
		return m.handleDashboardKey(msg)
	}
}

func (m *Model) handleDashboardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Quit):
		// Shutdown all services before quitting
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m.manager.Shutdown(ctx)
		return m, tea.Quit

	case key.Matches(msg, Keys.Up):
		if m.selected > 0 {
			m.selected--
		}

	case key.Matches(msg, Keys.Down):
		if m.selected < len(m.statuses)-1 {
			m.selected++
		}

	case key.Matches(msg, Keys.Start):
		return m, m.startSelected()

	case key.Matches(msg, Keys.Stop):
		return m, m.stopSelected()

	case key.Matches(msg, Keys.Restart):
		return m, m.restartSelected()

	case key.Matches(msg, Keys.StartAll):
		return m, m.startAll()

	case key.Matches(msg, Keys.StopAll):
		return m, m.stopAll()

	case key.Matches(msg, Keys.Kill):
		return m, m.killSelected()

	case key.Matches(msg, Keys.Logs):
		if len(m.statuses) > 0 && m.selected < len(m.statuses) {
			m.logsService = m.statuses[m.selected].Name
			m.viewMode = ViewLogs
			m.loadLogs()
		}
	}

	return m, nil
}

func (m *Model) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, Keys.Back):
		m.viewMode = ViewDashboard
		m.logEntries = nil
		m.logOffset = 0

	case key.Matches(msg, Keys.Up):
		if m.logOffset > 0 {
			m.logOffset--
		}

	case key.Matches(msg, Keys.Down):
		maxOffset := len(m.logEntries) - (m.height - 6)
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.logOffset < maxOffset {
			m.logOffset++
		}
	}

	return m, nil
}

func (m *Model) handleServiceEvent(event process.Event) {
	switch event.Type {
	case process.EventStarted:
		m.setMessage(fmt.Sprintf("✓ %s started", event.Service))
	case process.EventStopped:
		m.setMessage(fmt.Sprintf("○ %s stopped", event.Service))
	case process.EventCrashed:
		m.setMessage(fmt.Sprintf("✖ %s crashed", event.Service))
	}
	m.statuses = m.manager.AllStatus()
}

func (m *Model) setMessage(msg string) {
	m.message = msg
	m.msgExpiry = time.Now().Add(3 * time.Second)
}

// View renders the UI.
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	switch m.viewMode {
	case ViewLogs:
		return m.renderLogs()
	default:
		return m.renderDashboard()
	}
}

func (m *Model) renderDashboard() string {
	var b strings.Builder

	// Header
	header := TitleStyle.Render(" DevServ Dashboard ")
	timestamp := SubtitleStyle.Render(time.Now().Format("15:04:05"))
	headerLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		header,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(header)-lipgloss.Width(timestamp)-2)),
		timestamp,
	)
	b.WriteString(headerLine)
	b.WriteString("\n\n")

	// Table header
	tableHeader := fmt.Sprintf("  %-15s %-12s %-8s %-8s %-12s", "SERVICE", "STATUS", "PID", "PORT", "UPTIME")
	b.WriteString(SubtitleStyle.Render(tableHeader))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(strings.Repeat("─", min(m.width-2, 60))))
	b.WriteString("\n")

	// Services
	for i, status := range m.statuses {
		row := m.renderServiceRow(status, i == m.selected)
		b.WriteString(row)
		b.WriteString("\n")
	}

	if len(m.statuses) == 0 {
		b.WriteString(SubtitleStyle.Render("  No services configured"))
		b.WriteString("\n")
	}

	// Fill remaining space
	usedLines := 4 + len(m.statuses) + 4 // header + table + footer
	for i := 0; i < m.height-usedLines; i++ {
		b.WriteString("\n")
	}

	// Message line
	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(m.message)
	} else if m.err != nil {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render(m.err.Error()))
	} else {
		b.WriteString("\n")
	}

	// Footer with help
	b.WriteString("\n")
	helpLine := HelpStyle.Render("[s]tart [x]stop [r]estart [l]ogs [?]help [q]uit")
	b.WriteString(helpLine)

	return b.String()
}

func (m *Model) renderServiceRow(status process.Status, selected bool) string {
	// Status with symbol
	statusStyle := GetStatusStyle(status.State.String())
	statusText := fmt.Sprintf("%s %s", status.State.Symbol(), status.State.String())
	statusText = statusStyle.Render(statusText)

	// PID
	pid := "-"
	if status.PID > 0 {
		pid = fmt.Sprintf("%d", status.PID)
	}

	// Port
	port := "-"
	if status.Port > 0 {
		port = fmt.Sprintf("%d", status.Port)
	}

	// Uptime
	uptime := "-"
	if status.Uptime > 0 {
		uptime = formatDuration(status.Uptime)
	}

	// Build row
	selector := "  "
	if selected {
		selector = "▸ "
	}

	row := fmt.Sprintf("%s%-15s %-12s %-8s %-8s %-12s",
		selector, status.Name, statusText, pid, port, uptime)

	if selected {
		return SelectedStyle.Render(row)
	}
	return row
}

func (m *Model) renderLogs() string {
	var b strings.Builder

	// Header
	header := TitleStyle.Render(fmt.Sprintf(" Logs: %s ", m.logsService))
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("Press ESC to go back, ↑/↓ to scroll"))
	b.WriteString("\n\n")

	// Log content
	visibleLines := m.height - 6
	if visibleLines < 1 {
		visibleLines = 1
	}

	start := m.logOffset
	end := start + visibleLines
	if end > len(m.logEntries) {
		end = len(m.logEntries)
	}

	if len(m.logEntries) == 0 {
		b.WriteString(SubtitleStyle.Render("No logs available"))
	} else {
		for i := start; i < end; i++ {
			b.WriteString(m.logEntries[i])
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *Model) renderHelp() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(" Keyboard Shortcuts "))
	b.WriteString("\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Move up"},
		{"↓/j", "Move down"},
		{"s", "Start selected service"},
		{"x", "Stop selected service"},
		{"r", "Restart selected service"},
		{"S", "Start all services"},
		{"X", "Stop all services"},
		{"K", "Force kill selected service"},
		{"l/Enter", "View logs"},
		{"Esc", "Go back"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}

	for _, s := range shortcuts {
		b.WriteString(fmt.Sprintf("  %-12s %s\n", s.key, s.desc))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press ? or Esc to close"))

	return b.String()
}

func (m *Model) loadLogs() {
	m.logEntries = nil
	m.logOffset = 0

	logFile, err := m.manager.LogManager().GetLatestLog(m.logsService)
	if err != nil || logFile == nil {
		return
	}

	reader, err := m.manager.LogManager().CreateReader(logFile.Path, struct {
		Follow  bool
		Lines   int
		Since   time.Time
		Until   time.Time
		Pattern string
		Stream  string
	}{
		Lines: 100,
	})
	if err != nil {
		return
	}
	defer reader.Close()

	entries, err := reader.ReadAll()
	if err != nil {
		return
	}

	for _, entry := range entries {
		line := fmt.Sprintf("%s │ %s",
			LogTimestamp.Render(entry.Timestamp.Format("15:04:05")),
			entry.Message)
		m.logEntries = append(m.logEntries, line)
	}
}

// Commands

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *Model) startSelected() tea.Cmd {
	if len(m.statuses) == 0 || m.selected >= len(m.statuses) {
		return nil
	}

	name := m.statuses[m.selected].Name
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.manager.StartService(ctx, name); err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

func (m *Model) stopSelected() tea.Cmd {
	if len(m.statuses) == 0 || m.selected >= len(m.statuses) {
		return nil
	}

	name := m.statuses[m.selected].Name
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := m.manager.StopService(ctx, name); err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

func (m *Model) restartSelected() tea.Cmd {
	if len(m.statuses) == 0 || m.selected >= len(m.statuses) {
		return nil
	}

	name := m.statuses[m.selected].Name
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := m.manager.RestartService(ctx, name); err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

func (m *Model) killSelected() tea.Cmd {
	if len(m.statuses) == 0 || m.selected >= len(m.statuses) {
		return nil
	}

	name := m.statuses[m.selected].Name
	return func() tea.Msg {
		if err := m.manager.Kill(name); err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

func (m *Model) startAll() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.manager.Start(ctx); err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

func (m *Model) stopAll() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := m.manager.Stop(ctx); err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

// Helper functions

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
