# Makefile for Harlequin Remote Signing Service

.PHONY: build clean test deps help run dev docker-build docker-run docker-stop docker-clean

# Default target
help:
	@echo "🎭 Harlequin Remote Signing Service"
	@echo ""
	@echo "Available commands:"
	@echo "  build           Build the remote signing binary (current platform)"
	@echo "  clean           Clean build artifacts"
	@echo "  test            Run tests"
	@echo "  deps            Download dependencies"
	@echo "  docs            Generate API documentation"
	@echo "  run             Run the server (development)"
	@echo "  dev             Run in development mode with auto-reload"
	@echo ""
	@echo "Cross-compilation:"
	@echo "  build-linux     Build for Linux (amd64)"
	@echo "  build-linux-arm64 Build for Linux (arm64)"
	@echo "  build-darwin    Build for macOS (amd64)"
	@echo "  build-darwin-arm64 Build for macOS (arm64)"
	@echo "  build-windows   Build for Windows (amd64)"
	@echo "  build-all       Build for all platforms"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build         Build Docker image (with Go compilation)"
	@echo "  docker-build-binary  Build Docker image from pre-built Linux binary"
	@echo "  docker-build-arm64   Build ARM64 Docker image from pre-built binary"
	@echo "  docker-run           Run with docker compose"
	@echo "  docker-stop          Stop docker compose services"
	@echo "  docker-clean         Clean Docker images and containers"
	@echo ""
	@echo "  help         Show this help message"

# Build the binary
build:
	@echo "🔨 Building remote signing service..."
	go build -o remote-signing ./cmd/remote-signing
	@echo "✅ Build complete: ./remote-signing"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f remote-signing remote-signing-* harlequin-remote-signing harlequin-remote-signing-* test-build test-swagger
	rm -rf ../dist/remote-signing
	rm -f docs/swagger.json docs/swagger.yaml
	go clean
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies updated"

# Generate API documentation
docs:
	@echo "📚 Generating API documentation..."
	$(shell go env GOPATH)/bin/swag init -g cmd/remote-signing/main.go -o docs --parseDependency --parseInternal
	@echo "✅ API documentation generated in docs/"

# Run the server in development mode
run: build
	@echo "🚀 Starting remote signing server..."
	./remote-signing start

# Development mode with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@echo "🔄 Starting development server with auto-reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "❌ Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run 'make run' for normal mode"; \
	fi

# Install the binary to $GOPATH/bin
install: build
	@echo "📦 Installing remote signing service..."
	cp remote-signing $(GOPATH)/bin/
	@echo "✅ Installed to $(GOPATH)/bin/remote-signing"

# Cross-compilation targets
build-linux:
	@echo "🔨 Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o remote-signing-linux-amd64 ./cmd/remote-signing
	@echo "✅ Linux binary: ./remote-signing-linux-amd64"

build-linux-arm64:
	@echo "🔨 Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o remote-signing-linux-arm64 ./cmd/remote-signing
	@echo "✅ Linux ARM64 binary: ./remote-signing-linux-arm64"

build-darwin:
	@echo "🔨 Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o remote-signing-darwin-amd64 ./cmd/remote-signing
	@echo "✅ macOS binary: ./remote-signing-darwin-amd64"

build-darwin-arm64:
	@echo "🔨 Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o remote-signing-darwin-arm64 ./cmd/remote-signing
	@echo "✅ macOS ARM64 binary: ./remote-signing-darwin-arm64"

build-windows:
	@echo "🔨 Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o remote-signing-windows-amd64.exe ./cmd/remote-signing
	@echo "✅ Windows binary: ./remote-signing-windows-amd64.exe"

build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows
	@echo "✅ All binaries built"

# Docker commands
docker-build:
	@echo "🐳 Building Docker image (with Go compilation)..."
	docker build -t harlequin-remote-signing:latest .
	@echo "✅ Docker image built: harlequin-remote-signing:latest"

docker-build-binary: build-linux
	@echo "🐳 Building Docker image from pre-built binary..."
	cp remote-signing-linux-amd64 remote-signing
	docker build -f Dockerfile.binary -t harlequin-remote-signing:binary .
	rm remote-signing
	@echo "✅ Docker image built from binary: harlequin-remote-signing:binary"

docker-build-arm64: build-linux-arm64
	@echo "🐳 Building Docker image for ARM64..."
	cp remote-signing-linux-arm64 remote-signing
	docker build -f Dockerfile.binary -t harlequin-remote-signing:arm64 --platform linux/arm64 .
	rm remote-signing
	@echo "✅ Docker ARM64 image built: harlequin-remote-signing:arm64"

docker-run:
	@echo "🚀 Starting remote signing service with docker compose..."
	docker compose up -d
	@echo "✅ Service started. Check status with: docker compose ps"
	@echo "📡 Server available at: http://localhost:8080"

docker-stop:
	@echo "🛑 Stopping docker compose services..."
	docker compose down
	@echo "✅ Services stopped"

docker-clean:
	@echo "🧹 Cleaning Docker images and containers..."
	docker compose down --rmi all --volumes --remove-orphans
	docker image prune -f
	@echo "✅ Docker cleanup complete"

docker-logs:
	@echo "📋 Showing docker compose logs..."
	docker compose logs -f
