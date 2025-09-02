# Harlequin Toolkit Makefile

# Variables
BIN_DIR = bin
BINARY_NAME = harlequin
BINARY_PATH = $(BIN_DIR)/$(BINARY_NAME)

# Default target
.DEFAULT_GOAL := help

# Create bin directory if it doesn't exist
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# Build the CLI binary
.PHONY: build
build: $(BIN_DIR)
	@echo "Building harlequin CLI..."
	go build -o $(BINARY_PATH) .
	@echo "✅ Build complete: $(BINARY_PATH)"

# Development command: build and run
.PHONY: dev
dev: build
	@echo "Starting harlequin in development mode..."
	./$(BINARY_PATH)

# Run the CLI directly without building binary
.PHONY: run
run:
	@echo "Running harlequin CLI directly..."
	go run .

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	@echo "✅ Clean complete"

# Run all tests
.PHONY: test
test:
	@echo "Running all tests..."
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@$(MAKE) test-e2e

# Run unit tests (CLI internal tests)
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	go test ./...
	cd tests && go test ./unit/...

# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	cd tests && go test ./integration/...

# Run end-to-end tests
.PHONY: test-e2e
test-e2e:
	@echo "Running end-to-end tests..."
	cd tests && go test ./e2e/...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	cd tests && go test -cover ./...

# Run tests in verbose mode
.PHONY: test-verbose
test-verbose:
	@echo "Running tests in verbose mode..."
	go test -v ./...
	cd tests && go test -v ./...

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download

# Display help
.PHONY: help
help:
	@echo "Harlequin Toolkit - Available commands:"
	@echo ""
	@echo "  build              Build the harlequin CLI binary to bin/harlequin"
	@echo "  dev                Build and run the CLI binary"
	@echo "  run                Run the CLI directly with go run (no binary)"
	@echo "  clean              Remove build artifacts"
	@echo "  test               Run all tests (unit, integration, e2e)"
	@echo "  test-unit          Run unit tests only"
	@echo "  test-integration   Run integration tests only"
	@echo "  test-e2e           Run end-to-end tests only"
	@echo "  test-coverage      Run tests with coverage reporting"
	@echo "  test-verbose       Run tests in verbose mode"
	@echo "  deps               Install and tidy Go dependencies"
	@echo "  help               Show this help message"
	@echo ""
