#!/bin/bash

# Script to run the AOS integration test and generate permanent process.wasm for verification

echo "🚀 Running AOS Build Integration Test..."
echo "This will generate a permanent process.wasm file you can verify."
echo ""

cd "$(dirname "$0")"

# Run the integration test
go test -v -run TestBuildProjectWithInjection_Integration

echo ""
echo "📁 If the test succeeded, you should find the build outputs at:"
echo "   $(pwd)/test-output/output/"
echo ""
echo "🔍 Files to verify:"
echo "   - process.wasm: The compiled WebAssembly module"
echo "   - bundled.lua: The bundled Lua code that was injected"
echo ""
echo "You can inspect these files to verify the build process worked correctly!"
