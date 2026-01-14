# DevServ - Claude Code Guidelines

## Project Overview

DevServ is a CLI tool for managing multiple local development services with an interactive TUI dashboard, structured logging, and intelligent process management.

**Tech Stack:** Go, Cobra (CLI), Bubble Tea (TUI), Lip Gloss (styling), TOML (config)

## Project Structure

```
devserv/
├── cmd/devserv/main.go           # Entry point
├── internal/
│   ├── cli/                      # Cobra commands (start, stop, status, logs, ui, etc.)
│   ├── config/                   # TOML parsing and validation
│   ├── process/                  # Process manager, service lifecycle, state machine
│   ├── logs/                     # Log writer/reader with JSON formatting
│   └── tui/                      # Bubble Tea dashboard and views
├── docs/                         # Documentation
└── devserv.example.toml          # Example configuration
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/config/config.go` | Config structs and TOML parsing |
| `internal/process/manager.go` | Service orchestration |
| `internal/process/service.go` | Service lifecycle and process spawning |
| `internal/process/state.go` | State machine (stopped, starting, running, stopping, crashed) |
| `internal/logs/manager.go` | Log file organization and cleanup |
| `internal/logs/writer.go` | JSON log writing |
| `internal/logs/reader.go` | Log reading with filtering |
| `internal/tui/model.go` | Root Bubble Tea model |
| `internal/tui/views/` | Dashboard and log viewer components |

## Build & Test Commands

```bash
# Build
make build
# or
go build -o devserv ./cmd/devserv

# Run all tests
make test
# or
go test ./...

# Run with verbose output
go test -v ./internal/...

# Run specific test
go test -v -run TestE2EFullFlow ./internal/...

# Run the application
./devserv ui
./devserv start
./devserv status
```

## Architecture Patterns

### Process Management
- Services use a state machine: `stopped → starting → running → stopping → stopped/crashed`
- Graceful shutdown: SIGTERM with timeout, then SIGKILL
- Each service restart creates a new log file

### Log System
- Logs stored in `~/.devserv/logs/{service}/{YYYY-MM-DD}/{HH-MM-SS}.log`
- JSON structured format with timestamp, service, stream (stdout/stderr), message
- Reader supports filtering by pattern, stream, and time range

### TUI Architecture
- Bubble Tea with Elm architecture (Model, Update, View)
- Views: Dashboard (main), LogViewer, Help overlay
- Real-time updates via event channel from process manager

## Code Conventions

- Use `internal/` for all packages (not exported)
- Error handling: wrap errors with context using `fmt.Errorf("context: %w", err)`
- Logging in services goes through the log writer, not stdout
- Tests use real processes and file system (no mocks for E2E tests)
- Shell commands in tests should use script files, not inline `sh -c` (due to command parsing)

## Testing Notes

- Unit tests: config parsing, state machine, log formatting
- Integration tests: real process start/stop with actual commands
- E2E tests: full flow without mocks (config → start → logs → stop)
- Use `t.TempDir()` for temporary files in tests
- For commands with complex shell syntax, create temporary `.sh` scripts

## Common Tasks

### Adding a new CLI command
1. Create file in `internal/cli/`
2. Add command in `init()` with `rootCmd.AddCommand()`
3. Follow existing patterns (see `start.go`, `stop.go`)

### Adding a new TUI view
1. Create view in `internal/tui/views/`
2. Implement `Init()`, `Update()`, `View()` methods
3. Add navigation in main model

### Modifying service lifecycle
1. Update state machine in `internal/process/state.go`
2. Update service logic in `internal/process/service.go`
3. Add tests for new states/transitions
