# Getting Started with DevServ

This guide will walk you through setting up DevServ to manage your local development services.

## Prerequisites

- Go 1.21 or later
- macOS or Linux

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/marcelocajueiro/devserv@latest
```

### Option 2: Build from Source

```bash
git clone https://github.com/marcelocajueiro/devserv.git
cd devserv
go build -o devserv ./cmd/devserv

# Optionally move to PATH
sudo mv devserv /usr/local/bin/
```

## Your First Configuration

### 1. Create a Config File

Navigate to your project directory and run:

```bash
devserv init
```

This creates a `devserv.toml` file with example services.

### 2. Edit the Configuration

Open `devserv.toml` and configure your services:

```toml
# Backend API
[[services]]
name = "api"
command = "go run ./cmd/api"
directory = "./backend"
port = 8080

# Frontend Development Server
[[services]]
name = "frontend"
command = "npm run dev"
directory = "./frontend"
port = 3000

# Database (Docker)
[[services]]
name = "postgres"
command = "docker compose up postgres"
port = 5432
```

### 3. Start Your Services

Start all services:

```bash
devserv start
```

Or start specific services:

```bash
devserv start api frontend
```

## Using the Interactive Dashboard

Launch the TUI dashboard for a visual interface:

```bash
devserv ui
```

### Navigation

- **Arrow keys** or **j/k**: Move between services
- **s**: Start the selected service
- **x**: Stop the selected service
- **r**: Restart the selected service
- **l** or **Enter**: View logs
- **?**: Show help
- **q**: Quit

### Status Indicators

- `●` Green: Service is running
- `◐` Yellow: Service is starting/stopping
- `✖` Red: Service crashed
- `○` Gray: Service is stopped

## Viewing Logs

DevServ automatically captures stdout and stderr from your services.

### View Recent Logs

```bash
devserv logs api
```

### Follow Logs in Real-Time

```bash
devserv logs api -f
```

### Show More Lines

```bash
devserv logs api -n 500
```

## Managing Services

### Starting Services

```bash
# Start all services
devserv start

# Start specific services
devserv start api frontend
```

### Stopping Services

```bash
# Graceful stop (sends SIGTERM, waits 30s)
devserv stop api

# Force stop (sends SIGKILL immediately)
devserv stop -f api
```

### Restarting Services

```bash
devserv restart api
```

### Check Status

```bash
devserv status
```

Output:

```
SERVICE     STATUS       PID      PORT     UPTIME
-------     ------       ---      ----     ------
api         ● running    12345    8080     2h 15m
frontend    ● running    12346    3000     2h 15m
postgres    ○ stopped    -        5432     -
```

## Log Organization

Logs are stored in `~/.devserv/logs/` with the following structure:

```
~/.devserv/logs/
├── api/
│   └── 2024-01-15/
│       ├── 08-30-00.log
│       └── 14-20-15.log
├── frontend/
│   └── 2024-01-15/
│       └── 08-30-05.log
└── postgres/
    └── ...
```

Each service restart creates a new log file with a timestamp, making it easy to:

- Find logs from a specific session
- Debug issues from previous runs
- Keep logs organized by date

## Tips and Best Practices

### 1. Use Port Numbers

Always specify port numbers in your config. This allows DevServ to detect conflicts before starting:

```toml
[[services]]
name = "api"
command = "go run ./cmd/api"
port = 8080  # DevServ will warn if 8080 is in use
```

### 2. Organize Working Directories

Use the `directory` field to run services from their correct location:

```toml
[[services]]
name = "frontend"
command = "npm run dev"
directory = "./frontend"
```

### 3. Use the Dashboard for Development

The TUI dashboard (`devserv ui`) is ideal during active development:

- Real-time status updates
- Quick restart with `r`
- Easy log access with `l`

### 4. Use CLI for Scripts

Use CLI commands in your scripts or aliases:

```bash
# In your shell profile
alias dev="cd ~/projects/myapp && devserv start"
alias devui="cd ~/projects/myapp && devserv ui"
```

## Troubleshooting

### Service Won't Start

1. Check if the port is already in use:
   ```bash
   lsof -i :8080
   ```

2. Check the logs for errors:
   ```bash
   devserv logs api
   ```

### Service Crashes Immediately

1. Run the command manually to see the error:
   ```bash
   cd ./backend && go run ./cmd/api
   ```

2. Check if dependencies are installed (node_modules, go.mod, etc.)

### Dashboard Shows Wrong Status

The dashboard updates every second. If a service state seems stuck, it might be in the middle of starting or stopping.

## Next Steps

- [Configuration Reference](configuration.md) - All config options
- [CLI Reference](cli-reference.md) - Full command documentation
- [TUI Guide](tui-guide.md) - Dashboard features
- [Architecture](architecture.md) - How DevServ works
