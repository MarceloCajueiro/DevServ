.PHONY: build clean test run install lint fmt help

# Build variables
BINARY_NAME=devserv
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/MarceloCajueiro/DevServ/internal/cli.Version=$(VERSION) -X github.com/MarceloCajueiro/DevServ/internal/cli.GitCommit=$(GIT_COMMIT) -X github.com/MarceloCajueiro/DevServ/internal/cli.BuildDate=$(BUILD_DATE)"

# Default target
all: build

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/devserv

## install: Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/devserv

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

## test: Run tests
test:
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## run: Build and run
run: build
	./$(BINARY_NAME) ui

## tidy: Clean up dependencies
tidy:
	go mod tidy

## deps: Download dependencies
deps:
	go mod download

## release: Build release binaries for multiple platforms
release: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/devserv
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/devserv
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/devserv
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/devserv

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
