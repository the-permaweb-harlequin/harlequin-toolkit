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

# Run tests
.PHONY: test
test:
	@echo "Running CLI tests..."
	go test ./...

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
	@echo "  build    Build the harlequin CLI binary to bin/harlequin"
	@echo "  dev      Build and run the CLI binary"
	@echo "  run      Run the CLI directly with go run (no binary)"
	@echo "  clean    Remove build artifacts"
	@echo "  test     Run CLI tests"
	@echo "  deps     Install and tidy Go dependencies"
	@echo "  help     Show this help message"
	@echo ""
