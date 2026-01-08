#!/bin/bash
set -e

echo "Building WASM..."
cd "$(dirname "$0")"

# Build the WASM binary
GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd

# Copy the Go WASM support file (location varies by Go version)
WASM_EXEC=$(find "$(go env GOROOT)" -name "wasm_exec.js" 2>/dev/null | head -1)
cp "$WASM_EXEC" web/

echo "Build complete!"
echo ""
echo "To run, serve the web/ directory with a local HTTP server:"
echo "  cd web && python3 -m http.server 8080"
echo ""
echo "Then open http://localhost:8080"
