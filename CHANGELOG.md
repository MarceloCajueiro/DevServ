# Changelog

All notable changes to DevServ will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release
- TOML-based configuration (`devserv.toml`)
- CLI commands: `start`, `stop`, `restart`, `status`, `logs`, `ui`, `init`, `version`
- Interactive TUI dashboard with Bubble Tea
- Process management with graceful shutdown
- Port conflict detection before starting services
- Structured JSON logging with per-session log files
- Log organization by service/date/time
- Keyboard shortcuts for all operations
- Documentation: README, getting-started, configuration, cli-reference, tui-guide, architecture, contributing

### Features
- Start/stop/restart individual or all services
- Real-time status monitoring in TUI
- Service status indicators (running, starting, stopping, crashed, stopped)
- PID, port, and uptime display
- Log viewing and tailing
- Force kill option for unresponsive services

## [0.1.0] - TBD

- Initial public release
