#!/bin/bash

# Build script for TetoDB WebAssembly module

set -e

echo "Building TetoDB WebAssembly module..."

# Build the WASM binary
GOOS=js GOARCH=wasm go build -o nodejs/wasm/tetodb.wasm ./wasm

echo "Build complete! WASM module at: nodejs/wasm/tetodb.wasm"
echo ""

# Copy the Go WASM runtime
echo "Copying wasm_exec.js..."
GOROOT=$(go env GOROOT)
if [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/misc/wasm/wasm_exec.js" nodejs/wasm/
elif [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/lib/wasm/wasm_exec.js" nodejs/wasm/
else
    echo "Error: wasm_exec.js not found in GOROOT"
    exit 1
fi

echo ""
echo "Build successful!"
echo ""
echo "Next steps:"
echo "  1. cd nodejs"
echo "  2. npm install"
echo "  3. node src/server.js"
