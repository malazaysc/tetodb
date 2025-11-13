package engine

import (
	"fmt"
	"sync"
)

// Database represents the main database instance
// It manages multiple collections and coordinates persistence
type Database struct {
	storage     *Storage                // Underlying storage layer
	collections map[string]*Collection  // Map of collection name -> Collection
	mu          sync.RWMutex            // Protects access to collections map
}

// OpenDatabase opens (or creates) a database at the given file path
// It loads all existing data from the file into memory
func OpenDatabase(path string) (*Database, error) {
	// Create storage layer
	storage, err := NewStorage(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	db := &Database{
		storage:     storage,
		collections: make(map[string]*Collection),
	}

	// Load all records from disk
	if err := db.loadFromDisk(); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to load from disk: %w", err)
	}

	return db, nil
}

// loadFromDisk reads all records from storage and rebuilds the in-memory collections
func (db *Database) loadFromDisk() error {
	records, err := db.storage.LoadAll()
	if err != nil {
		return err
	}

	// Reconstruct collections from records
	// We use a temporary map to track the latest version of each document
	tempData := make(map[string]map[string]map[string]interface{})

	for _, record := range records {
		// Ensure collection exists in temp map
		if tempData[record.Collection] == nil {
			tempData[record.Collection] = make(map[string]map[string]interface{})
		}

		// If doc is nil, it means this document was deleted
		if record.Doc == nil {
			delete(tempData[record.Collection], record.ID)
		} else {
			// Store or update the document
			tempData[record.Collection][record.ID] = record.Doc
		}
	}

	// Create Collection objects from the temp data
	for collName, docs := range tempData {
		if len(docs) > 0 {
			coll := NewCollection(collName, db.storage)
			coll.documents = docs
			db.collections[collName] = coll
		}
	}

	return nil
}

// GetCollection returns a collection by name
// Creates the collection if it doesn't exist
func (db *Database) GetCollection(name string) *Collection {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if collection already exists
	if coll, exists := db.collections[name]; exists {
		return coll
	}

	// Create new collection
	coll := NewCollection(name, db.storage)
	db.collections[name] = coll
	return coll
}

// ListCollections returns a list of all collection names
func (db *Database) ListCollections() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	names := make([]string, 0, len(db.collections))
	for name := range db.collections {
		names = append(names, name)
	}
	return names
}

// DropCollection removes a collection and all its documents
func (db *Database) DropCollection(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	coll, exists := db.collections[name]
	if !exists {
		return nil // Collection doesn't exist, nothing to do
	}

	// Delete all documents in the collection
	for id := range coll.documents {
		if err := coll.Delete(id); err != nil {
			return fmt.Errorf("failed to delete document: %w", err)
		}
	}

	// Remove collection from map
	delete(db.collections, name)
	return nil
}

// Close closes the database and flushes all data to disk
func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.storage != nil {
		return db.storage.Close()
	}
	return nil
}

// Compact performs compaction on the storage file
// This removes deleted/updated records and reclaims disk space
func (db *Database) Compact() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Collect all current records
	var records []StorageRecord
	for collName, coll := range db.collections {
		for id, doc := range coll.documents {
			records = append(records, StorageRecord{
				Collection: collName,
				ID:         id,
				Doc:        doc,
			})
		}
	}

	return db.storage.Compact(records)
}

// Stats returns statistics about the database
func (db *Database) Stats() map[string]interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	stats := map[string]interface{}{
		"collections": len(db.collections),
		"documents":   0,
	}

	totalDocs := 0
	collStats := make(map[string]int)
	for name, coll := range db.collections {
		count := len(coll.documents)
		collStats[name] = count
		totalDocs += count
	}

	stats["documents"] = totalDocs
	stats["collection_stats"] = collStats

	return stats
}
