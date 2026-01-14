# DevServ

**Local development server manager for developers who run multiple services.**

DevServ simplifies managing your local development stack with an interactive TUI dashboard, structured logging, and intelligent process management.

## Features

- **Interactive Dashboard** - Real-time service status, CPU, memory, uptime
- **Simple Configuration** - TOML-based config with minimal boilerplate
- **Structured Logs** - JSON logs organized by service, date, and session
- **Process Management** - Start, stop, restart with graceful shutdown
- **Port Detection** - Automatically detects port conflicts before starting
- **Keyboard-Driven** - Navigate and control services without leaving the terminal

## Quick Start

```bash
# Install
go install github.com/marcelocajueiro/devserv/cmd/devserv@v0.2.0

# Create config
devserv init

# Edit devserv.toml with your services
# Then start everything
devserv start

# Or launch the interactive dashboard
devserv ui
```

## Configuration

DevServ looks for configuration in this order:
1. `./devserv.toml` (current directory)
2. `~/.config/devserv/config.toml`
3. `~/.devserv/config.toml` (global fallback)

You can also specify a config file with `-c /path/to/config.toml`.

### Local Config (per-project)

Create a `devserv.toml` file in your project:

```toml
[[services]]
name = "api"
command = "go run ./cmd/api"
directory = "./backend"
port = 8080

[[services]]
name = "frontend"
command = "npm run dev"
directory = "./frontend"
port = 3000

[[services]]
name = "postgres"
command = "docker compose up postgres"
port = 5432
```

### Global Config

Create `~/.devserv/config.toml` for services you want available everywhere:

```toml
[[services]]
name = "redis"
command = "redis-server"
port = 6379

[[services]]
name = "mailhog"
command = "mailhog"
port = 8025
```

Now you can run `devserv start redis` from any directory.

### Service Options

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique service identifier |
| `command` | Yes | Command to run the service |
| `directory` | No | Working directory (default: current) |
| `port` | No | Port number (for conflict detection) |

## CLI Commands

```bash
devserv start [service...]   # Start services
devserv stop [service...]    # Stop services (graceful)
devserv stop -f [service...] # Force kill services
devserv restart [service...] # Restart services
devserv status               # Show status table
devserv logs <service> [-f]  # View/tail service logs
devserv ui                   # Interactive dashboard
devserv init                 # Generate example config
devserv version              # Show version
```

## Interactive Dashboard

Launch with `devserv ui`:

```
┌─────────────────────────────────────────────────────────────────┐
│  DevServ Dashboard                                     12:34:56 │
├─────────────────────────────────────────────────────────────────┤
│  SERVICE     STATUS       PID    PORT   UPTIME                  │
│  ───────────────────────────────────────────────────────────────│
│▸ api         ● running    1234   8080   2h 15m                  │
│  frontend    ● running    1235   3000   2h 15m                  │
│  postgres    ○ stopped    -      5432   -                       │
├─────────────────────────────────────────────────────────────────┤
│  [s]tart [x]stop [r]estart [l]ogs [?]help [q]uit               │
└─────────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/k` | Move selection up |
| `↓/j` | Move selection down |
| `s` | Start selected service |
| `x` | Stop selected service |
| `r` | Restart selected service |
| `S` | Start all services |
| `X` | Stop all services |
| `K` | Force kill selected |
| `l` / `Enter` | View logs |
| `Esc` | Go back |
| `?` | Show help |
| `q` | Quit |

## Log Management

Logs are stored in `~/.devserv/logs/` with the following structure:

```
~/.devserv/logs/
├── api/
│   └── 2024-01-15/
│       ├── 08-30-00.log    # Session started at 08:30
│       └── 14-20-15.log    # Restart at 14:20
└── frontend/
    └── ...
```

Each log file contains JSON-structured entries:

```json
{"timestamp":"2024-01-15T08:30:00Z","service":"api","stream":"stdout","message":"Server started on :8080"}
```

View logs with:

```bash
devserv logs api          # Show last 100 lines
devserv logs api -n 50    # Show last 50 lines
devserv logs api -f       # Follow (tail) logs
```

## Installation

### From Source

```bash
go install github.com/marcelocajueiro/devserv/cmd/devserv@v0.2.0
```

### Build from Repository

```bash
git clone https://github.com/marcelocajueiro/devserv.git
cd devserv
go build -o devserv ./cmd/devserv
```

## Documentation

- [Getting Started](docs/getting-started.md) - Detailed tutorial
- [Configuration Reference](docs/configuration.md) - All config options
- [CLI Reference](docs/cli-reference.md) - Full command documentation
- [TUI Guide](docs/tui-guide.md) - Dashboard usage guide
- [Architecture](docs/architecture.md) - How DevServ works internally
- [Contributing](docs/contributing.md) - Development guide

## Requirements

- Go 1.21 or later
- macOS or Linux (Windows support planned)

## License

MIT License - see [LICENSE](LICENSE)

## Contributing

Contributions are welcome! Please read our [Contributing Guide](docs/contributing.md) before submitting PRs.
