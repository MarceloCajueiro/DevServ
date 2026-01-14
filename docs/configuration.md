# Configuration Reference

DevServ uses TOML for configuration. This document covers all available options.

## Config File Location

DevServ looks for configuration in this order:

1. Path specified with `--config` flag
2. `./devserv.toml` (current directory)
3. `~/.config/devserv/config.toml`
4. `~/.devserv/config.toml`

## Basic Structure

```toml
# Each service is defined in a [[services]] block
[[services]]
name = "service-name"
command = "command to run"
directory = "./optional/path"
port = 8080
```

## Service Options

### Required Fields

#### `name`

Unique identifier for the service.

```toml
[[services]]
name = "api"
```

- Must be unique across all services
- Used in CLI commands: `devserv start api`
- Used in log directory names

#### `command`

The command to execute to start the service.

```toml
[[services]]
command = "go run ./cmd/api"
```

- Can include arguments
- Executed using the system shell
- Environment variables from your shell are inherited

### Optional Fields

#### `directory`

Working directory for the service.

```toml
[[services]]
directory = "./backend"
```

- Default: Current directory where `devserv` is run
- Supports relative and absolute paths
- Relative paths are resolved from where `devserv` is executed

#### `port`

Port number the service listens on.

```toml
[[services]]
port = 8080
```

- Used for port conflict detection before starting
- DevServ will error if the port is already in use
- Optional but recommended

## Complete Example

```toml
# DevServ Configuration
# https://github.com/marcelocajueiro/devserv

# Go API Backend
[[services]]
name = "api"
command = "go run ./cmd/api"
directory = "./backend"
port = 8080

# React Frontend
[[services]]
name = "frontend"
command = "npm run dev"
directory = "./frontend"
port = 3000

# Background Worker
[[services]]
name = "worker"
command = "python -m worker"
directory = "./worker"

# PostgreSQL via Docker Compose
[[services]]
name = "postgres"
command = "docker compose up postgres"
port = 5432

# Redis via Docker Compose
[[services]]
name = "redis"
command = "docker compose up redis"
port = 6379

# Celery Worker (no port, background process)
[[services]]
name = "celery"
command = "celery -A tasks worker"
directory = "./backend"
```

## Common Patterns

### Docker Compose Services

```toml
[[services]]
name = "postgres"
command = "docker compose up postgres"
port = 5432
```

For multiple Docker services, you can either:

1. Define each as a separate service (recommended for independent control)
2. Use a single service that runs `docker compose up`

### Node.js Applications

```toml
[[services]]
name = "frontend"
command = "npm run dev"
directory = "./frontend"
port = 3000
```

### Go Applications

```toml
[[services]]
name = "api"
command = "go run ./cmd/api"
directory = "./backend"
port = 8080
```

Or with air for hot reload:

```toml
[[services]]
name = "api"
command = "air"
directory = "./backend"
port = 8080
```

### Python Applications

```toml
# Flask
[[services]]
name = "api"
command = "flask run"
directory = "./backend"
port = 5000

# Django
[[services]]
name = "api"
command = "python manage.py runserver"
directory = "./backend"
port = 8000

# FastAPI with uvicorn
[[services]]
name = "api"
command = "uvicorn main:app --reload"
directory = "./backend"
port = 8000
```

### Background Workers

Services without ports (background workers, queue consumers):

```toml
[[services]]
name = "worker"
command = "node worker.js"
directory = "./worker"
# No port - this is a background worker
```

## Tips

### Environment Variables

Environment variables from your shell are inherited. You can also set them inline:

```toml
[[services]]
name = "api"
command = "DATABASE_URL=postgres://localhost/db go run ./cmd/api"
```

Or use a wrapper script:

```toml
[[services]]
name = "api"
command = "./scripts/start-api.sh"
```

### Multiple Commands

For complex startup sequences, use a shell script:

```toml
[[services]]
name = "api"
command = "bash -c 'source .env && go run ./cmd/api'"
```

### Virtual Environments (Python)

```toml
[[services]]
name = "api"
command = "bash -c 'source venv/bin/activate && python app.py'"
directory = "./backend"
```

Or activate in the command:

```toml
[[services]]
name = "api"
command = "./venv/bin/python app.py"
directory = "./backend"
```

## Validation

DevServ validates your configuration on startup:

- **name**: Must be present and unique
- **command**: Must be present and non-empty
- **port**: If specified, must be available when starting

If validation fails, DevServ will show a clear error message:

```
Error: invalid configuration: service "api": command is required
```
