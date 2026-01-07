#!/bin/bash
set -e

echo "Building WASM..."
cd "$(dirname "$0")"

# Build the WASM binary
GOOS=js GOARCH=wasm go build -o wasm/main.wasm ./wasm

# Copy the Go WASM support file
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" wasm/

echo "Build complete!"
echo ""
echo "To run, serve the wasm/ directory with a local HTTP server:"
echo "  cd wasm && python3 -m http.server 8080"
echo ""
echo "Then open http://localhost:8080"
