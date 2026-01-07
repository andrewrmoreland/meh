#!/bin/bash
set -e

echo "Building WASM..."
cd "$(dirname "$0")"

# Build the WASM binary
GOOS=js GOARCH=wasm go build -o wasm/main.wasm ./wasm

# Copy the Go WASM support file (location varies by Go version)
WASM_EXEC=$(find "$(go env GOROOT)" -name "wasm_exec.js" 2>/dev/null | head -1)
cp "$WASM_EXEC" wasm/

echo "Build complete!"
echo ""
echo "To run, serve the wasm/ directory with a local HTTP server:"
echo "  cd wasm && python3 -m http.server 8080"
echo ""
echo "Then open http://localhost:8080"
