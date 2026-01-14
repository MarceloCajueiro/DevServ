# TUI Guide

The interactive terminal user interface (TUI) provides a visual dashboard for managing your services.

## Launching the Dashboard

```bash
devserv ui
```

## Dashboard Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  DevServ Dashboard                                     12:34:56 │
├─────────────────────────────────────────────────────────────────┤
│  SERVICE     STATUS       PID    PORT   UPTIME                  │
│  ───────────────────────────────────────────────────────────────│
│▸ api         ● running    1234   8080   2h 15m                  │
│  frontend    ● running    1235   3000   2h 15m                  │
│  postgres    ○ stopped    -      5432   -                       │
│  worker      ✖ crashed    -      -      -                       │
├─────────────────────────────────────────────────────────────────┤
│  [s]tart [x]stop [r]estart [l]ogs [?]help [q]uit               │
└─────────────────────────────────────────────────────────────────┘
```

### Components

1. **Header** - Title and current time
2. **Service Table** - List of all configured services
3. **Status Bar** - Keyboard shortcuts hint

### Status Indicators

| Symbol | Color | Meaning |
|--------|-------|---------|
| `●` | Green | Service is running |
| `◐` | Yellow | Service is starting |
| `◑` | Yellow | Service is stopping |
| `✖` | Red | Service crashed |
| `○` | Gray | Service is stopped |

### Table Columns

- **SERVICE**: Service name from config
- **STATUS**: Current state with indicator
- **PID**: Process ID (when running)
- **PORT**: Configured port number
- **UPTIME**: Time since started

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑` or `k` | Move selection up |
| `↓` or `j` | Move selection down |

### Service Control

| Key | Action |
|-----|--------|
| `s` | Start selected service |
| `x` | Stop selected service (graceful) |
| `r` | Restart selected service |
| `K` | Force kill selected (SIGKILL) |

### Bulk Operations

| Key | Action |
|-----|--------|
| `S` | Start all services |
| `X` | Stop all services |

### Views

| Key | Action |
|-----|--------|
| `l` or `Enter` | View logs for selected |
| `Esc` | Go back to dashboard |
| `?` | Toggle help overlay |

### Application

| Key | Action |
|-----|--------|
| `q` | Quit (stops all services) |
| `Ctrl+C` | Quit (stops all services) |

## Log Viewer

Press `l` or `Enter` on a service to view its logs:

```
┌─────────────────────────────────────────────────────────────────┐
│  Logs: api                                                      │
│  Press ESC to go back, ↑/↓ to scroll                           │
├─────────────────────────────────────────────────────────────────┤
│  12:34:56 │ Server starting on :8080                            │
│  12:34:57 │ Connected to database                               │
│  12:34:58 │ Ready to accept connections                         │
│  12:35:01 │ GET /api/users 200 15ms                             │
│  12:35:02 │ GET /api/posts 200 23ms                             │
└─────────────────────────────────────────────────────────────────┘
```

### Log Viewer Controls

| Key | Action |
|-----|--------|
| `↑` or `k` | Scroll up |
| `↓` or `j` | Scroll down |
| `Esc` | Return to dashboard |

## Help Overlay

Press `?` to see all keyboard shortcuts:

```
┌─────────────────────────────────────────────────────────────────┐
│  Keyboard Shortcuts                                             │
├─────────────────────────────────────────────────────────────────┤
│  ↑/k          Move up                                           │
│  ↓/j          Move down                                         │
│  s            Start selected service                            │
│  x            Stop selected service                             │
│  r            Restart selected service                          │
│  S            Start all services                                │
│  X            Stop all services                                 │
│  K            Force kill selected service                       │
│  l/Enter      View logs                                         │
│  Esc          Go back                                           │
│  ?            Toggle help                                       │
│  q            Quit                                              │
│                                                                 │
│  Press ? or Esc to close                                        │
└─────────────────────────────────────────────────────────────────┘
```

## Status Messages

The dashboard shows temporary status messages when actions complete:

- `✓ api started` - Service started successfully
- `○ api stopped` - Service stopped
- `✖ api crashed` - Service exited unexpectedly

Messages disappear after 3 seconds.

## Tips

### Quick Restart Workflow

1. Select service with `↓` or `↑`
2. Press `r` to restart
3. Watch status change from `stopping` → `starting` → `running`

### Debugging a Crash

1. Select the crashed service (`✖ crashed`)
2. Press `l` to view logs
3. Scroll up to find error messages
4. Press `Esc` to return
5. Press `s` to try starting again

### Starting Fresh

1. Press `X` to stop all services
2. Press `S` to start all services

### Exiting Gracefully

Press `q` to quit. DevServ will:
1. Send SIGTERM to all running services
2. Wait for graceful shutdown
3. Exit cleanly

If services don't stop in time, they'll be force-killed.

## Troubleshooting

### Dashboard Doesn't Fit

The dashboard adapts to terminal size. If it looks wrong:
- Resize your terminal to be wider
- Minimum recommended: 80 columns × 24 rows

### Keys Not Working

- Make sure your terminal supports escape sequences
- Try a different terminal emulator if issues persist

### Status Not Updating

The dashboard updates every second. If a service state seems stuck:
- It might be in the middle of starting/stopping
- Check logs for any errors
