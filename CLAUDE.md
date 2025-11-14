# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TetoDB is a tiny embeddable NoSQL database engine written in Go, compiled to WebAssembly, and designed for use in Node.js applications. It's a document-oriented database similar to SQLite but for JSON documents (like MongoDB), providing a simple file-based storage engine.

**Important**: This is a learning/educational project, not production-ready.

## Build Commands

### Building the WebAssembly Module

```bash
# Using Docker (no Go installation required)
make docker-build

# Using local Go installation
make build

# Alternative using build script
./build.sh

# Manual build (if needed)
GOOS=js GOARCH=wasm go build -o nodejs/wasm/tetodb.wasm ./wasm
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" nodejs/wasm/
```

### Running Tests

```bash
# Go tests for the engine
go test ./engine/... -v

# Node.js tests (not yet implemented)
cd nodejs && npm test
```

### Running the Demo Server

```bash
# Using Docker (no Go installation required)
make docker-run

# Using local Go installation
make run

# Manual steps
make build  # or make docker-build
cd nodejs
npm install
npm start

# Development mode with auto-reload
cd nodejs && npm run dev
```

### Other Useful Commands

```bash
make clean    # Remove build artifacts and .db files
make install  # Install Node.js dependencies
make help     # Show all available make commands
```

## Architecture

### Three-Layer Design

1. **Go Engine Layer** (`engine/`): Core database implementation
   - `storage.go`: Low-level file I/O, append-only log format, compaction
   - `db.go`: Database instance, manages collections, startup/loading, stats
   - `collection.go`: CRUD operations on collections
   - `query.go`: Document filtering and matching logic

2. **WASM Bridge Layer** (`wasm/main.go`): Exposes Go functions to JavaScript
   - Global database instance management
   - JavaScript function registration (tetoDBOpen, tetoDBInsert, etc.)
   - JSON serialization/deserialization between JS and Go
   - Error handling and result formatting

3. **JavaScript Wrapper Layer** (`nodejs/src/tetodb.js`): Promise-based Node.js API
   - TetoDB class: Database instance with open/close/stats/compact methods
   - Collection class: Document operations (insert, find, update, delete, count)
   - WASM module initialization and lifecycle management

### Storage Format

- **Append-only log**: Each line is a JSON record: `{"collection": "name", "id": "uuid", "doc": {...}}`
- **Updates**: Append new version of document (old version remains until compaction)
- **Deletes**: Append record with `"doc": null`
- **On startup**: Read entire file, build in-memory map `collection -> id -> document`
- **Compaction**: Rewrite file with only current document versions

### Key Design Decisions

- **Single file storage**: All collections stored in one database file (like SQLite)
- **In-memory index**: Entire database loaded into memory for fast reads (limitation: not suitable for large datasets)
- **No concurrency**: Single-threaded, no locking or ACID guarantees
- **Simple queries**: Only equality matching (e.g., `{name: "Alice", role: "admin"}` with AND logic)
- **UUID-based IDs**: Using github.com/google/uuid for document IDs

## Common Development Patterns

### Adding New Query Operations

If extending query capabilities beyond equality matching:
1. Modify `engine/query.go` to add new matching logic
2. Update `Collection.Find()` in `engine/collection.go` to use new logic
3. No changes needed to WASM layer or JS wrapper (they pass filters as JSON)

### Adding New Database Operations

To add a new database-level operation:
1. Implement in `engine/db.go` (e.g., new method on Database struct)
2. Export in `wasm/main.go` (register with `js.Global().Set()`)
3. Add wrapper method in `nodejs/src/tetodb.js` TetoDB class

### Adding New Collection Operations

To add a new collection-level operation:
1. Implement in `engine/collection.go` (e.g., new method on Collection struct)
2. Export in `wasm/main.go` with collection name + other args
3. Add wrapper method in `nodejs/src/tetodb.js` Collection class

### Error Handling Pattern

- Go layer: Return Go errors from functions
- WASM layer: Convert to `makeError()` response objects
- JS layer: Check `result.success` and throw Error if false

## File Locations

- Built WASM files: `nodejs/wasm/` (generated, not committed)
  - `tetodb.wasm`: Compiled Go code
  - `wasm_exec.js`: Go WASM runtime (copied from GOROOT)
- Demo server: `nodejs/src/server.js` (Express-based REST API)
- Go dependencies: `go.mod` (currently only github.com/google/uuid)
- Node dependencies: `nodejs/package.json`

## Prerequisites

- **Go 1.23+** OR **Docker** (for building WASM - choose one)
- **Node.js 18+** (required for running the application)
- **Make** (optional, for using Makefile)

## Storage Architecture

TetoDB uses different storage implementations based on the build target:

- **WASM Build** (`engine/storage_memory.go`): Uses a JavaScript-to-Go bridge to access Node.js file system operations. File persistence works via `nodeFileReadSync`, `nodeFileWriteSync`, `nodeFileAppendSync`, etc. exposed from `nodejs/src/tetodb.js`.
- **Native Go Build** (`engine/storage_file.go`): Uses standard Go file I/O with `os.OpenFile`, `bufio.Scanner`, etc.

Both implementations provide the same append-only log format for compatibility.

## Known Limitations

- No transactions or ACID guarantees
- No concurrency control (single-threaded)
- Simple queries only (no $gt, $lt, $in, regex, etc.)
- No secondary indexes (all queries scan the collection)
- Not optimized for large datasets (entire DB in memory)
- No schema validation
