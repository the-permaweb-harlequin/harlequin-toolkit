.PHONY: build build-debug test clean help

# Default target
build: build/process.wasm

# Build release version
build/process.wasm: main.go go.mod
	@mkdir -p build
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o build/process.wasm main.go
	@echo "✅ Built release version: build/process.wasm"

# Build debug version
build-debug: build/process-debug.wasm
build/process-debug.wasm: main.go go.mod
	@mkdir -p build
	GOOS=js GOARCH=wasm go build -o build/process-debug.wasm main.go
	@echo "✅ Built debug version: build/process-debug.wasm"

# Test the process
test: build/process.wasm
	@echo "🧪 Testing AO Go process..."
	@if command -v tsx >/dev/null 2>&1; then \
		tsx --test --noEmit test/test.ts; \
	elif command -v node >/dev/null 2>&1; then \
		node --test --loader ts-node/esm test/test.ts; \
	else \
		echo "❌ Node.js with tsx or ts-node is required for testing"; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	rm -rf build/
	@echo "🧹 Cleaned build directory"

# Development workflow
dev: clean build test

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build release version (default)"
	@echo "  build-debug - Build debug version with symbols"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  dev         - Clean, build, and test"
	@echo "  help        - Show this help"
