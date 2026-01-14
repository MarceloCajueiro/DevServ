# Contributing to DevServ

Thank you for your interest in contributing to DevServ! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- A terminal with good Unicode support (for TUI testing)

### Clone and Build

```bash
git clone https://github.com/MarceloCajueiro/DevServ.git
cd devserv
go mod download
go build -o devserv ./cmd/devserv
```

### Run Tests

```bash
go test ./...
```

### Run Locally

```bash
# Create a test config
./devserv init

# Start services
./devserv start

# Or run the TUI
./devserv ui
```

## Project Structure

```
devserv/
├── cmd/devserv/         # Entry point
├── internal/
│   ├── cli/             # CLI commands (Cobra)
│   ├── config/          # Configuration
│   ├── process/         # Process management
│   ├── logs/            # Log management
│   └── tui/             # Terminal UI (Bubble Tea)
├── docs/                # Documentation
└── README.md
```

See [Architecture](architecture.md) for detailed component descriptions.

## Making Changes

### 1. Create a Branch

```bash
git checkout -b feat/my-feature
# or
git checkout -b fix/my-bugfix
```

### 2. Make Your Changes

- Follow existing code style
- Add tests for new functionality
- Update documentation if needed

### 3. Test Your Changes

```bash
# Run tests
go test ./...

# Build and test manually
go build -o devserv ./cmd/devserv
./devserv init
./devserv start
```

### 4. Commit Your Changes

Use [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Features
git commit -m "feat: add service health checks"

# Bug fixes
git commit -m "fix: correct port conflict detection"

# Documentation
git commit -m "docs: update configuration reference"

# Refactoring
git commit -m "refactor: simplify process manager"
```

### 5. Create a Pull Request

1. Push your branch
2. Open a PR against `main`
3. Fill out the PR template
4. Wait for review

## Code Style

### Go Code

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

### Commit Messages

- Use conventional commits format
- Write clear, descriptive messages
- Reference issues when applicable

### Documentation

- Update docs when changing user-facing behavior
- Use clear, concise language
- Include examples where helpful

## Areas for Contribution

### Good First Issues

- Documentation improvements
- Error message clarity
- Test coverage improvements
- Small bug fixes

### Feature Ideas

- Service dependencies (start order)
- Health check support
- Web-based dashboard
- Plugin system
- Windows support

### Performance

- Reduce memory usage
- Optimize log reading
- Improve startup time

## Testing

### Unit Tests

```bash
go test ./internal/config/
go test ./internal/process/
go test ./internal/logs/
```

### Integration Tests

```bash
go test ./... -tags=integration
```

### Manual Testing

1. Create a test project with multiple services
2. Test all CLI commands
3. Test the TUI thoroughly
4. Test edge cases (crashes, port conflicts, etc.)

## Pull Request Guidelines

### Before Submitting

- [ ] Code builds without errors
- [ ] All tests pass
- [ ] New code has tests
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventional commits

### PR Description

Include:
- What the change does
- Why it's needed
- How to test it
- Any breaking changes

### Review Process

1. Maintainers will review your PR
2. Address any feedback
3. Once approved, it will be merged

## Getting Help

- Open an issue for bugs or feature requests
- Ask questions in discussions
- Check existing issues and PRs first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
