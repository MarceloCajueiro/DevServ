# Architecture

This document explains how DevServ works internally.

## Overview

DevServ is built with a modular architecture:

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Interface                          │
│                                                                 │
│    ┌─────────────────┐           ┌─────────────────┐           │
│    │   CLI Commands  │           │   TUI Dashboard │           │
│    │  (Cobra-based)  │           │  (Bubble Tea)   │           │
│    └────────┬────────┘           └────────┬────────┘           │
└─────────────┼────────────────────────────┼──────────────────────┘
              │                            │
              ▼                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Process Manager                            │
│                                                                 │
│    ┌─────────────────────────────────────────────────────┐     │
│    │  Manager                                             │     │
│    │  - Service orchestration                            │     │
│    │  - Event distribution                               │     │
│    │  - Port conflict detection                          │     │
│    └──────────────────────┬──────────────────────────────┘     │
│                           │                                     │
│    ┌──────────────────────┼──────────────────────────────┐     │
│    │                      ▼                              │     │
│    │  ┌──────────┐  ┌──────────┐  ┌──────────┐         │     │
│    │  │ Service  │  │ Service  │  │ Service  │  ...    │     │
│    │  │ "api"    │  │"frontend"│  │"postgres"│         │     │
│    │  └────┬─────┘  └────┬─────┘  └────┬─────┘         │     │
│    │       │             │             │                │     │
│    │       ▼             ▼             ▼                │     │
│    │  [os/exec]     [os/exec]     [os/exec]            │     │
│    └────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Log Manager                              │
│                                                                 │
│    ┌──────────────────────────────────────────────────────┐    │
│    │  ~/.devserv/logs/                                    │    │
│    │  ├── api/                                            │    │
│    │  │   └── 2024-01-15/                                │    │
│    │  │       └── 08-30-00.log                           │    │
│    │  ├── frontend/                                       │    │
│    │  └── postgres/                                       │    │
│    └──────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
devserv/
├── cmd/devserv/
│   └── main.go              # Entry point
├── internal/
│   ├── cli/                 # CLI commands (Cobra)
│   │   ├── root.go          # Root command and config loading
│   │   ├── start.go         # Start command
│   │   ├── stop.go          # Stop command
│   │   ├── restart.go       # Restart command
│   │   ├── status.go        # Status command
│   │   ├── logs.go          # Logs command
│   │   ├── ui.go            # UI command (launches TUI)
│   │   ├── init.go          # Init command
│   │   └── version.go       # Version command
│   ├── config/              # Configuration
│   │   └── config.go        # TOML parsing and validation
│   ├── process/             # Process management
│   │   ├── manager.go       # Service orchestration
│   │   ├── service.go       # Individual service lifecycle
│   │   └── state.go         # Service state definitions
│   ├── logs/                # Log management
│   │   ├── manager.go       # Log file organization
│   │   ├── writer.go        # JSON log writing
│   │   └── reader.go        # Log reading and tailing
│   └── tui/                 # Terminal UI
│       ├── tui.go           # TUI entry point
│       ├── model.go         # Bubble Tea model
│       ├── keys.go          # Keyboard shortcuts
│       └── styles.go        # Lip Gloss styles
└── docs/                    # Documentation
```

## Core Components

### Config (`internal/config/`)

Handles TOML configuration parsing:

```go
type Config struct {
    Services []Service
}

type Service struct {
    Name      string
    Command   string
    Directory string
    Port      int
}
```

Key functions:
- `Load(path)` - Load and validate config
- `Validate()` - Check config validity
- `GetService(name)` - Find service by name

### Process Manager (`internal/process/`)

#### Manager

Orchestrates all services:

```go
type Manager struct {
    config     *config.Config
    services   map[string]*Service
    logManager *logs.Manager
    eventsCh   chan Event
}
```

Key methods:
- `Start(ctx, names...)` - Start services
- `Stop(ctx, names...)` - Stop services
- `Restart(ctx, names...)` - Restart services
- `AllStatus()` - Get all service statuses
- `Events()` - Subscribe to events

#### Service

Manages individual process lifecycle:

```go
type Service struct {
    config    config.Service
    state     State
    cmd       *exec.Cmd
    pid       int
    startTime time.Time
    logWriter *logs.Writer
}
```

States:
- `StateStopped` - Not running
- `StateStarting` - Being started
- `StateRunning` - Running normally
- `StateStopping` - Shutting down
- `StateCrashed` - Exited unexpectedly

### Log Manager (`internal/logs/`)

#### Directory Structure

```
~/.devserv/logs/
├── {service}/
│   └── {YYYY-MM-DD}/
│       └── {HH-MM-SS}.log
```

#### Log Entry Format

JSON-structured entries:

```json
{
  "timestamp": "2024-01-15T08:30:00.123Z",
  "service": "api",
  "stream": "stdout",
  "message": "Server started on :8080"
}
```

#### Writer

Creates new log file per service session:

```go
writer, err := logManager.CreateWriter("api")
writer.WriteLine("stdout", "message")
```

#### Reader

Reads and optionally follows logs:

```go
reader, err := logManager.CreateReader(path, ReaderOptions{
    Follow: true,
    Lines:  100,
})
for entry := range reader.Read(ctx) {
    // Process entry
}
```

### TUI (`internal/tui/`)

Built with Bubble Tea (Elm architecture):

```go
type Model struct {
    manager  *process.Manager
    viewMode ViewMode  // Dashboard, Logs, Help
    selected int       // Selected service index
    statuses []process.Status
}

func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

Key messages:
- `tea.KeyMsg` - Keyboard input
- `tea.WindowSizeMsg` - Terminal resize
- `TickMsg` - Periodic refresh
- `ServiceEventMsg` - Service state changes

## Data Flow

### Starting a Service

```
1. User runs: devserv start api
                    │
                    ▼
2. CLI loads config, creates Manager
                    │
                    ▼
3. Manager.StartService("api")
                    │
                    ├── Check port availability
                    │
                    ├── Create log writer
                    │
                    ▼
4. Service.Start()
                    │
                    ├── Parse command
                    │
                    ├── exec.Command()
                    │
                    ├── Start goroutines:
                    │   - captureOutput (stdout)
                    │   - captureOutput (stderr)
                    │   - waitForExit
                    │
                    ▼
5. Service running, state = StateRunning
```

### Stopping a Service

```
1. User presses Ctrl+C or runs: devserv stop api
                    │
                    ▼
2. Manager.StopService("api")
                    │
                    ▼
3. Service.Stop(ctx)
                    │
                    ├── Send SIGTERM
                    │
                    ├── Wait for exit or timeout (30s)
                    │
                    ├── If timeout: Send SIGKILL
                    │
                    ▼
4. waitForExit goroutine detects exit
                    │
                    ├── Close log writer
                    │
                    ├── Update state = StateStopped
                    │
                    ├── Send EventStopped
                    │
                    ▼
5. Service stopped
```

### Event Flow (TUI)

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Service    │────▶│   Manager    │────▶│  Event Chan  │
│  (goroutine) │     │              │     │              │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                 │
                          ┌──────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │   TUI Event Loop      │
              │                       │
              │  - Update statuses    │
              │  - Show notifications │
              │  - Re-render view     │
              └───────────────────────┘
```

## Key Design Decisions

### Why Go?

- Fast startup time (important for CLI tools)
- Single binary distribution
- Excellent concurrency primitives (goroutines for output capture)
- Strong ecosystem (Cobra, Bubble Tea)

### Why Bubble Tea?

- Elm architecture (predictable state management)
- Pure functions for rendering (testable)
- Active ecosystem (Lip Gloss for styling)
- Excellent terminal compatibility

### Why TOML?

- Human-readable
- Comment support
- Less error-prone than YAML
- Popular in developer tools (Cargo, Poetry)

### Why JSON Logs?

- Structured and searchable
- Easy to parse programmatically
- Standard format for log aggregation
- Preserves all metadata

### Why Per-Session Log Files?

- Easy to find logs from specific runs
- No complex rotation logic
- Natural organization by time
- Simple cleanup (delete old directories)

## Extending DevServ

### Adding a New CLI Command

1. Create `internal/cli/mycommand.go`:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    RunE:  runMyCommand,
}

func init() {
    // Add to root in root.go
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

2. Add to `rootCmd` in `internal/cli/root.go`

### Adding a TUI View

1. Create constants for the view mode
2. Add view state to Model
3. Handle view in `handleKey()`
4. Add render function
5. Switch in `View()`

### Adding Service Features

New service options:

1. Add field to `config.Service`
2. Use in `process.Service.Start()`
3. Document in configuration.md
