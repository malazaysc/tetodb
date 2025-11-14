# TetoDB

A tiny embeddable NoSQL database engine written in Go, compiled to WebAssembly, and designed to be used from Node.js applications.

TetoDB is similar to SQLite but document-oriented (like MongoDB). It provides a simple, file-based storage engine with support for multiple collections, basic CRUD operations, and simple querying capabilities.

## Features

- **Document-Oriented**: Store JSON documents in named collections
- **File-Based Storage**: All data stored in a single file (like SQLite)
- **Simple API**: Clean, Promise-based JavaScript API
- **WebAssembly**: Written in Go, compiled to WASM for portability
- **No External Dependencies**: Embedded database - no server needed
- **Basic Query Support**: Filter documents by field values
- **In-Memory Indexing**: Fast lookups with in-memory document cache
- **Compaction**: Reclaim disk space from deleted/updated records

## Architecture

```
┌─────────────────────────────────────────┐
│         Node.js Application             │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │   Express API (server.js)        │  │
│  └───────────────┬───────────────────┘  │
│                  │                      │
│  ┌───────────────▼───────────────────┐  │
│  │   JS Wrapper (tetodb.js)       │  │
│  └───────────────┬───────────────────┘  │
│                  │                      │
│  ┌───────────────▼───────────────────┐  │
│  │   WebAssembly Module              │  │
│  │   (tetodb.wasm)                 │  │
│  │                                   │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │  Go Database Engine         │  │  │
│  │  │  - Collection Management    │  │  │
│  │  │  - Query Engine             │  │  │
│  │  │  - Storage Layer            │  │  │
│  │  └─────────────┬───────────────┘  │  │
│  └────────────────┼───────────────────┘  │
└───────────────────┼───────────────────────┘
                    │
                    ▼
            ┌───────────────┐
            │  File System  │
            │  (data.db)    │
            └───────────────┘
```

## Project Structure

```
tetodb/
├── engine/              # Go database engine
│   ├── storage.go      # File-based storage layer
│   ├── db.go           # Database management
│   ├── collection.go   # Collection operations
│   └── query.go        # Query and filtering logic
├── wasm/               # WebAssembly entry point
│   └── main.go         # WASM exports and JS bindings
├── nodejs/             # Node.js integration
│   ├── src/
│   │   ├── tetodb.js # JavaScript wrapper API
│   │   └── server.js   # Express demo server
│   ├── wasm/           # Built WASM files (generated)
│   │   ├── tetodb.wasm
│   │   └── wasm_exec.js
│   └── package.json
├── go.mod              # Go dependencies
├── Makefile            # Build automation
├── build.sh            # Build script
└── README.md           # This file
```

## Prerequisites

- **Go 1.23+** OR **Docker**: Required to build the WebAssembly module (choose one)
- **Node.js 18+**: Required to run the demo application
- **Make** (optional): For using the Makefile

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd tetodb
```

### 2. Build the WebAssembly Module

**Option A: Using Docker (no Go installation required)**
```bash
make docker-build
```

**Option B: Using Make (requires Go)**
```bash
make build
```

**Option C: Using the Build Script (requires Go)**
```bash
./build.sh
```

**Option D: Manual Build (requires Go)**
```bash
# Download Go dependencies
go mod download

# Build the WASM module
GOOS=js GOARCH=wasm go build -o nodejs/wasm/tetodb.wasm ./wasm

# Copy the Go WASM runtime
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" nodejs/wasm/
```

### 3. Install Node.js Dependencies

```bash
cd nodejs
npm install
```

### 4. Run the Demo Server

```bash
npm start
```

The server will start on `http://localhost:3000`

## Using the Demo API

Once the server is running, you can interact with it using curl or any HTTP client.

### Create a User

```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "age": 30, "role": "admin"}'
```

Response:
```json
{
  "success": true,
  "message": "User created successfully",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "role": "admin",
    "createdAt": "2024-01-15T10:30:00.000Z",
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### List All Users

```bash
curl http://localhost:3000/users
```

### Get User by ID

```bash
curl http://localhost:3000/users/550e8400-e29b-41d4-a716-446655440000
```

### Update a User

```bash
curl -X PATCH http://localhost:3000/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"age": 31}'
```

### Delete a User

```bash
curl -X DELETE http://localhost:3000/users/550e8400-e29b-41d4-a716-446655440000
```

### Get Database Statistics

```bash
curl http://localhost:3000/stats
```

### Compact the Database

```bash
curl -X POST http://localhost:3000/compact
```

## JavaScript API Reference

### Opening a Database

```javascript
const { TetoDB } = require('./src/tetodb');

const db = new TetoDB();
await db.open('mydata.db');
```

### Working with Collections

```javascript
// Get a collection
const users = db.collection('users');

// Insert a document
const id = await users.insert({
  name: 'Alice',
  email: 'alice@example.com',
  age: 25
});

// Find documents
const allUsers = await users.find();
const adults = await users.find({ age: 25 });

// Find by ID
const user = await users.findById(id);

// Update a document
await users.updateById(id, { age: 26 });

// Delete a document
await users.deleteById(id);

// Count documents
const count = await users.count();
const adultCount = await users.count({ age: 26 });
```

### Database Operations

```javascript
// Get statistics
const stats = await db.stats();
console.log(stats);
// { collections: 2, documents: 150, collection_stats: { users: 100, posts: 50 } }

// Compact the database
await db.compact();

// Close the database
await db.close();
```

## How It Works

### Storage Format

TetoDB uses a simple append-only log format:

- Each line in the file is a JSON-encoded record
- Format: `{"collection": "users", "id": "123", "doc": {...}}`
- Updates append a new version of the document
- Deletes append a record with `"doc": null`
- Compaction removes old versions and reclaims space

### In-Memory Index

On startup, TetoDB:
1. Reads all records from the file
2. Builds an in-memory map: `collection -> id -> document`
3. Newer records override older ones
4. Records with `null` documents are deletions

This provides fast reads while maintaining durability.

### Query Engine

The query engine supports simple equality filters:

```javascript
// Find all users named "Alice"
await users.find({ name: "Alice" });

// Find all admin users
await users.find({ role: "admin" });

// Multiple conditions (AND logic)
await users.find({ role: "admin", status: "active" });
```

## Limitations

This is a **learning project** and **not production-ready**. Known limitations:

- **In-Memory Only (WASM)**: When compiled to WebAssembly, TetoDB runs entirely in memory with no file persistence. Data is lost when the process stops. (Native Go builds support file persistence.)
- **No Transactions**: No ACID guarantees
- **No Concurrency**: Single-threaded, no locking
- **Simple Queries**: Only equality matching (no $gt, $lt, $in, etc.)
- **No Indexes**: All queries scan the collection
- **Limited Performance**: Not optimized for large datasets
- **No Schema Validation**: Documents can have any structure
- **Memory Usage**: Entire database loaded into memory

## Future Enhancements

Possible improvements for learning:

- [ ] Advanced query operators ($gt, $lt, $in, $regex)
- [ ] Secondary indexes for faster queries
- [ ] Pagination support
- [ ] Bulk operations
- [ ] Transactions (MVCC)
- [ ] Schema validation
- [ ] Aggregation pipeline
- [ ] Full-text search
- [ ] Replication
- [ ] Encryption at rest

## Troubleshooting

### Build Fails

**Problem**: `go: module not found`

**Solution**: Run `go mod download` or `go mod tidy`

### WASM Module Not Loading

**Problem**: `Cannot find module '../wasm/tetodb.wasm'`

**Solution**: Make sure you've built the WASM module first:
```bash
make build
```

### Server Won't Start

**Problem**: `Error: Database is not open`

**Solution**: Ensure the WASM module was built correctly and `wasm_exec.js` is in `nodejs/wasm/`

### Port Already in Use

**Problem**: `Error: listen EADDRINUSE: address already in use :::3000`

**Solution**: Change the port:
```bash
PORT=3001 npm start
```

## Development

### Running Tests

```bash
# Go tests
go test ./engine/... -v

# Node.js tests (not implemented yet)
cd nodejs
npm test
```

### Building for Different Architectures

The WASM module is architecture-independent, but you can build native Go binaries:

```bash
# For your current platform
go build -o tetodb ./wasm

# For specific platforms
GOOS=linux GOARCH=amd64 go build -o tetodb-linux ./wasm
GOOS=darwin GOARCH=arm64 go build -o tetodb-mac ./wasm
GOOS=windows GOARCH=amd64 go build -o tetodb.exe ./wasm
```

### Code Organization

- **engine/storage.go**: Low-level file I/O and record serialization
- **engine/db.go**: Database instance, collection management, startup logic
- **engine/collection.go**: Collection operations (CRUD)
- **engine/query.go**: Query filtering and matching logic
- **wasm/main.go**: WASM exports, JavaScript bindings
- **nodejs/src/tetodb.js**: JavaScript wrapper, Promise-based API
- **nodejs/src/server.js**: Express demo application

## Contributing

This is an educational project! Feel free to:

- Report bugs
- Suggest features
- Submit pull requests
- Use as a learning resource

## License

MIT License - See LICENSE file for details

## Acknowledgments

- Inspired by SQLite's single-file design
- Uses Go's WebAssembly support
- Built with Express.js for the demo

## Learn More

- [Go WebAssembly Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [WebAssembly.org](https://webassembly.org/)
- [Express.js](https://expressjs.com/)
- [SQLite Design](https://www.sqlite.org/arch.html)
- [MongoDB Concepts](https://www.mongodb.com/docs/)

---

**Happy Coding!**
