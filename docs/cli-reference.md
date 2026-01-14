# CLI Reference

Complete reference for all DevServ commands.

## Global Options

```bash
devserv [command] [flags]

Flags:
  -c, --config string   Config file path (default: ./devserv.toml)
  -h, --help           Help for any command
```

## Commands

### `devserv start`

Start one or more services.

```bash
devserv start [service...] [flags]
```

**Arguments:**
- `service...` - Optional. Service names to start. If omitted, starts all services.

**Examples:**
```bash
# Start all services
devserv start

# Start specific services
devserv start api
devserv start api frontend

# Start with custom config
devserv start -c ./config/dev.toml
```

**Behavior:**
- Checks port availability before starting
- Captures stdout/stderr to log files
- Blocks and waits for Ctrl+C to stop
- Gracefully stops all services on exit

---

### `devserv stop`

Stop one or more services.

```bash
devserv stop [service...] [flags]
```

**Arguments:**
- `service...` - Optional. Service names to stop. If omitted, stops all services.

**Flags:**
- `-f, --force` - Force kill services (SIGKILL instead of SIGTERM)

**Examples:**
```bash
# Graceful stop (SIGTERM, wait 30s)
devserv stop

# Stop specific service
devserv stop api

# Force kill (SIGKILL)
devserv stop -f api
```

**Behavior:**
- Sends SIGTERM for graceful shutdown
- Waits up to 30 seconds for process to exit
- Sends SIGKILL if process doesn't exit
- With `-f`, sends SIGKILL immediately

---

### `devserv restart`

Restart one or more services.

```bash
devserv restart [service...] [flags]
```

**Arguments:**
- `service...` - Optional. Service names to restart. If omitted, restarts all services.

**Examples:**
```bash
# Restart all services
devserv restart

# Restart specific service
devserv restart api
```

**Behavior:**
- Stops the service gracefully
- Waits briefly
- Starts the service again

---

### `devserv status`

Show status of all services.

```bash
devserv status
```

**Output:**
```
SERVICE     STATUS       PID      PORT     UPTIME
-------     ------       ---      ----     ------
api         ● running    12345    8080     2h 15m
frontend    ● running    12346    3000     2h 15m
postgres    ○ stopped    -        5432     -
```

**Status Indicators:**
- `● running` - Service is running normally
- `◐ starting` - Service is starting up
- `◑ stopping` - Service is shutting down
- `✖ crashed` - Service exited unexpectedly
- `○ stopped` - Service is not running

---

### `devserv logs`

View logs for a service.

```bash
devserv logs [service] [flags]
```

**Arguments:**
- `service` - Required. Service name to view logs for.

**Flags:**
- `-f, --follow` - Follow log output (tail)
- `-n, --lines int` - Number of lines to show (default: 100)

**Examples:**
```bash
# Show last 100 lines
devserv logs api

# Follow logs in real-time
devserv logs api -f

# Show last 50 lines
devserv logs api -n 50

# Show last 500 lines and follow
devserv logs api -n 500 -f
```

**Output:**
```
Logs for api (~/.devserv/logs/api/2024-01-15/08-30-00.log)
Press Ctrl+C to stop following

12:34:56.123 │ Server starting on :8080
12:34:56.456 │ Connected to database
12:34:57.789 │ Ready to accept connections
```

---

### `devserv ui`

Launch the interactive TUI dashboard.

```bash
devserv ui
```

**Features:**
- Real-time service status
- Start/stop/restart controls
- Log viewer
- Keyboard navigation

See [TUI Guide](tui-guide.md) for detailed usage.

---

### `devserv init`

Create a new configuration file.

```bash
devserv init
```

**Behavior:**
- Creates `devserv.toml` in current directory
- Fails if file already exists
- Includes example service definitions

**Output:**
```
Created /path/to/devserv.toml

Next steps:
  1. Edit devserv.toml to configure your services
  2. Run 'devserv start' to start all services
  3. Run 'devserv ui' for the interactive dashboard
```

---

### `devserv version`

Show version information.

```bash
devserv version
```

**Output:**
```
devserv v1.0.0
  commit: abc1234
  built:  2024-01-15
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (config not found, service not found, etc.) |

## Environment Variables

DevServ inherits environment variables from your shell. Services receive these variables when started.

## Signal Handling

DevServ handles the following signals:

- **SIGINT** (Ctrl+C): Initiates graceful shutdown
- **SIGTERM**: Initiates graceful shutdown

When shutting down:
1. Sends SIGTERM to all running services
2. Waits up to 30 seconds for each service
3. Sends SIGKILL to any remaining services
4. Exits
