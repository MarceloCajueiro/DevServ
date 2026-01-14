# Changelog

All notable changes to DevServ will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2026-01-14

### Added
- Initial "loading" state for services before state synchronization completes

### Fixed
- Verify process liveness when restoring "running" state from state file
- Services no longer briefly show as "running" on startup if process is dead
- Stale state files from crashed instances are automatically cleaned up

## [0.2.0] - 2026-01-14

### Added
- Shared state between multiple instances via file-based synchronization
- Real-time sync indicator showing last synchronization time in TUI
- Support for running multiple `devserv ui` instances simultaneously

### Fixed
- Service crash/stop state now properly propagates to all running instances
- State synchronization between instances updates every second
- Child processes (e.g., `uv run python ...`) now properly killed using process groups

## [0.1.0] - 2026-01-14

### Added
- TOML-based configuration (`devserv.toml`)
- CLI commands: `start`, `stop`, `restart`, `status`, `logs`, `ui`, `init`, `version`
- Interactive TUI dashboard with Bubble Tea
- Process management with graceful shutdown (SIGTERM → SIGKILL)
- Port conflict detection before starting services
- Structured JSON logging with per-session log files
- Log organization by service/date/time
- Log filtering by pattern and stream
- Keyboard shortcuts for all operations
- Force kill option for unresponsive services
- Comprehensive test suite (unit, integration, E2E)

### Documentation
- README with quick start guide
- Getting started tutorial
- Configuration reference
- CLI command reference
- TUI guide with keyboard shortcuts
- Architecture overview
- Contributing guide
