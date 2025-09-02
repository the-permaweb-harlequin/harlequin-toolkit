.PHONY: wasm test clean realclean

# Default target
wasm: src/process.wasm

# Build the Go WASM binary for AO compatibility  
src/process.wasm: main.go | src
	GOOS=js GOARCH=wasm go build -o src/process.wasm \
		-ldflags="-s -w" \
		.

src:
	mkdir -p src

# Test target that matches the parent project structure
test: wasm
	cd ../test && node --experimental-wasm-memory64 fixtures/test.ts ../ao-go-process/src/process.wasm wasm32-unknown-emscripten3 hello

# Alternative test that uses the Go binary directly
test-go: wasm
	@echo "Testing Go WASM binary..."
	@echo "Note: This requires adapting the test for Go WASM specifics"

clean:
	rm -f src/process.wasm

realclean: clean
	rm -rf src

# Development helpers
build-debug: main.go | src
	GOOS=js GOARCH=wasm go build -o src/process.wasm \
		-gcflags="-N -l" \
		.

# Show WASM info
info: src/process.wasm
	wasm-objdump -h src/process.wasm
	@echo "Size: $$(wc -c < src/process.wasm) bytes"

# Validate WASM
validate: src/process.wasm
	wasm-validate src/process.wasm
