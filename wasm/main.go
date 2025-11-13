package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/malazaysc/tetodb/engine"
)

// Global database instance
var db *engine.Database

// main is the entry point for the WASM module
// It registers JavaScript functions and keeps the Go runtime alive
func main() {
	fmt.Println("TetoDB WASM module loaded")

	// Register JavaScript functions
	js.Global().Set("tetoDBOpen", js.FuncOf(openDatabase))
	js.Global().Set("tetoDBInsert", js.FuncOf(insertDocument))
	js.Global().Set("tetoDBFind", js.FuncOf(findDocuments))
	js.Global().Set("tetoDBFindByID", js.FuncOf(findDocumentByID))
	js.Global().Set("tetoDBUpdate", js.FuncOf(updateDocument))
	js.Global().Set("tetoDBDelete", js.FuncOf(deleteDocument))
	js.Global().Set("tetoDBCount", js.FuncOf(countDocuments))
	js.Global().Set("tetoDBStats", js.FuncOf(getStats))
	js.Global().Set("tetoDBCompact", js.FuncOf(compactDatabase))
	js.Global().Set("tetoDBClose", js.FuncOf(closeDatabase))

	fmt.Println("TetoDB API functions registered")

	// Keep the Go runtime alive
	select {}
}

// openDatabase opens a database file
// Args: [path string]
// Returns: {success: bool, error: string}
func openDatabase(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return makeError("missing path argument")
	}

	path := args[0].String()

	var err error
	db, err = engine.OpenDatabase(path)
	if err != nil {
		return makeError(fmt.Sprintf("failed to open database: %v", err))
	}

	return makeSuccess(map[string]interface{}{
		"message": "Database opened successfully",
		"path":    path,
	})
}

// insertDocument inserts a document into a collection
// Args: [collection string, jsonDoc string]
// Returns: {success: bool, id: string, error: string}
func insertDocument(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if len(args) < 2 {
		return makeError("missing arguments: collection, jsonDoc")
	}

	collectionName := args[0].String()
	jsonDoc := args[1].String()

	// Parse JSON document
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(jsonDoc), &doc); err != nil {
		return makeError(fmt.Sprintf("invalid JSON: %v", err))
	}

	// Get collection
	coll := db.GetCollection(collectionName)

	// Insert document
	id, err := coll.Insert(doc)
	if err != nil {
		return makeError(fmt.Sprintf("insert failed: %v", err))
	}

	return makeSuccess(map[string]interface{}{
		"id": id,
	})
}

// findDocuments finds documents in a collection
// Args: [collection string, filterJSON string]
// Returns: {success: bool, documents: string (JSON array), error: string}
func findDocuments(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if len(args) < 1 {
		return makeError("missing collection argument")
	}

	collectionName := args[0].String()

	// Parse filter if provided
	var filter map[string]interface{}
	if len(args) >= 2 && args[1].String() != "" {
		if err := json.Unmarshal([]byte(args[1].String()), &filter); err != nil {
			return makeError(fmt.Sprintf("invalid filter JSON: %v", err))
		}
	}

	// Get collection
	coll := db.GetCollection(collectionName)

	// Find documents
	var docs []map[string]interface{}
	if len(filter) > 0 {
		docs = coll.Find(filter)
	} else {
		docs = coll.FindAll()
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(docs)
	if err != nil {
		return makeError(fmt.Sprintf("failed to serialize results: %v", err))
	}

	return makeSuccess(map[string]interface{}{
		"documents": string(jsonBytes),
		"count":     len(docs),
	})
}

// findDocumentByID finds a single document by ID
// Args: [collection string, id string]
// Returns: {success: bool, document: string (JSON), error: string}
func findDocumentByID(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if len(args) < 2 {
		return makeError("missing arguments: collection, id")
	}

	collectionName := args[0].String()
	id := args[1].String()

	// Get collection
	coll := db.GetCollection(collectionName)

	// Find document
	doc := coll.FindByID(id)
	if doc == nil {
		return makeError("document not found")
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return makeError(fmt.Sprintf("failed to serialize document: %v", err))
	}

	return makeSuccess(map[string]interface{}{
		"document": string(jsonBytes),
	})
}

// updateDocument updates a document in a collection
// Args: [collection string, id string, updateJSON string]
// Returns: {success: bool, error: string}
func updateDocument(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if len(args) < 3 {
		return makeError("missing arguments: collection, id, updateJSON")
	}

	collectionName := args[0].String()
	id := args[1].String()
	updateJSON := args[2].String()

	// Parse update JSON
	var update map[string]interface{}
	if err := json.Unmarshal([]byte(updateJSON), &update); err != nil {
		return makeError(fmt.Sprintf("invalid update JSON: %v", err))
	}

	// Get collection
	coll := db.GetCollection(collectionName)

	// Update document
	if err := coll.Update(id, update); err != nil {
		return makeError(fmt.Sprintf("update failed: %v", err))
	}

	return makeSuccess(map[string]interface{}{
		"message": "Document updated successfully",
	})
}

// deleteDocument deletes a document from a collection
// Args: [collection string, id string]
// Returns: {success: bool, error: string}
func deleteDocument(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if len(args) < 2 {
		return makeError("missing arguments: collection, id")
	}

	collectionName := args[0].String()
	id := args[1].String()

	// Get collection
	coll := db.GetCollection(collectionName)

	// Delete document
	if err := coll.Delete(id); err != nil {
		return makeError(fmt.Sprintf("delete failed: %v", err))
	}

	return makeSuccess(map[string]interface{}{
		"message": "Document deleted successfully",
	})
}

// countDocuments counts documents in a collection
// Args: [collection string, filterJSON string (optional)]
// Returns: {success: bool, count: int, error: string}
func countDocuments(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if len(args) < 1 {
		return makeError("missing collection argument")
	}

	collectionName := args[0].String()

	// Parse filter if provided
	var filter map[string]interface{}
	if len(args) >= 2 && args[1].String() != "" {
		if err := json.Unmarshal([]byte(args[1].String()), &filter); err != nil {
			return makeError(fmt.Sprintf("invalid filter JSON: %v", err))
		}
	}

	// Get collection
	coll := db.GetCollection(collectionName)

	// Count documents
	var count int
	if len(filter) > 0 {
		count = coll.CountWhere(filter)
	} else {
		count = coll.Count()
	}

	return makeSuccess(map[string]interface{}{
		"count": count,
	})
}

// getStats returns database statistics
// Args: []
// Returns: {success: bool, stats: object, error: string}
func getStats(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	stats := db.Stats()

	return makeSuccess(map[string]interface{}{
		"stats": stats,
	})
}

// compactDatabase performs database compaction
// Args: []
// Returns: {success: bool, error: string}
func compactDatabase(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if err := db.Compact(); err != nil {
		return makeError(fmt.Sprintf("compaction failed: %v", err))
	}

	return makeSuccess(map[string]interface{}{
		"message": "Database compacted successfully",
	})
}

// closeDatabase closes the database
// Args: []
// Returns: {success: bool, error: string}
func closeDatabase(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return makeError("database not open")
	}

	if err := db.Close(); err != nil {
		return makeError(fmt.Sprintf("close failed: %v", err))
	}

	db = nil

	return makeSuccess(map[string]interface{}{
		"message": "Database closed successfully",
	})
}

// makeSuccess creates a success response object
func makeSuccess(data map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"success": true,
	}

	for key, value := range data {
		result[key] = value
	}

	return result
}

// makeError creates an error response object
func makeError(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   message,
	}
}
