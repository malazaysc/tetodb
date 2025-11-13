.PHONY: all build clean test install run

# Build the WebAssembly module
all: build

build:
	@echo "Building MiniLiteDB WebAssembly module..."
	GOOS=js GOARCH=wasm go build -o nodejs/wasm/minilite.wasm ./wasm
	@echo "Build complete! WASM module at: nodejs/wasm/minilite.wasm"
	@echo ""
	@echo "Copying wasm_exec.js..."
	@if [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" nodejs/wasm/; \
	elif [ -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" nodejs/wasm/; \
	else \
		echo "Error: wasm_exec.js not found"; \
		exit 1; \
	fi
	@echo "Done!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f nodejs/wasm/minilite.wasm
	rm -f nodejs/wasm/wasm_exec.js
	rm -f *.db
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running Go tests..."
	go test ./engine/... -v

# Install Node.js dependencies
install:
	@echo "Installing Node.js dependencies..."
	cd nodejs && npm install
	@echo "Installation complete!"

# Build and run the demo server
run: build install
	@echo "Starting demo server..."
	cd nodejs && node src/server.js

# Build for different platforms (future expansion)
build-all: build

# Help command
help:
	@echo "MiniLiteDB Build Commands:"
	@echo "  make build    - Build the WebAssembly module"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run Go tests"
	@echo "  make install  - Install Node.js dependencies"
	@echo "  make run      - Build and run the demo server"
	@echo "  make help     - Show this help message"
