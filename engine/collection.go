package engine

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Collection represents a named collection of documents
// Similar to a table in SQL or a collection in MongoDB
type Collection struct {
	name      string                            // Collection name
	documents map[string]map[string]interface{} // Map of document ID -> document data
	storage   *Storage                          // Reference to storage layer
	mu        sync.RWMutex                      // Protects concurrent access to documents
}

// NewCollection creates a new Collection instance
func NewCollection(name string, storage *Storage) *Collection {
	return &Collection{
		name:      name,
		documents: make(map[string]map[string]interface{}),
		storage:   storage,
	}
}

// Insert adds a new document to the collection
// If the document doesn't have an "id" field, one is generated automatically
// Returns the document ID
func (c *Collection) Insert(doc map[string]interface{}) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if document has an ID, if not generate one
	var id string
	if idVal, exists := doc["id"]; exists {
		id = fmt.Sprintf("%v", idVal)
	} else {
		// Generate a new UUID
		id = uuid.New().String()
		doc["id"] = id
	}

	// Check if document with this ID already exists
	if _, exists := c.documents[id]; exists {
		return "", fmt.Errorf("document with id %s already exists", id)
	}

	// Store document in memory
	c.documents[id] = doc

	// Persist to disk
	record := StorageRecord{
		Collection: c.name,
		ID:         id,
		Doc:        doc,
	}

	if err := c.storage.Append(record); err != nil {
		// Rollback in-memory change if disk write fails
		delete(c.documents, id)
		return "", fmt.Errorf("failed to persist document: %w", err)
	}

	return id, nil
}

// FindByID retrieves a single document by its ID
// Returns nil if document doesn't exist
func (c *Collection) FindByID(id string) map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.documents[id]
}

// FindAll returns all documents in the collection
func (c *Collection) FindAll() []map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	docs := make([]map[string]interface{}, 0, len(c.documents))
	for _, doc := range c.documents {
		docs = append(docs, doc)
	}
	return docs
}

// Find searches for documents matching the given filter
// The filter is applied using the Query engine
func (c *Collection) Find(filter map[string]interface{}) []map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(filter) == 0 {
		// No filter, return all documents
		return c.FindAll()
	}

	var results []map[string]interface{}
	for _, doc := range c.documents {
		if MatchesFilter(doc, filter) {
			results = append(results, doc)
		}
	}

	return results
}

// Update modifies an existing document
// Merges the update fields into the existing document
func (c *Collection) Update(id string, update map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if document exists
	existingDoc, exists := c.documents[id]
	if !exists {
		return fmt.Errorf("document with id %s not found", id)
	}

	// Merge update into existing document
	for key, value := range update {
		existingDoc[key] = value
	}

	// Ensure ID is preserved
	existingDoc["id"] = id

	// Persist to disk
	record := StorageRecord{
		Collection: c.name,
		ID:         id,
		Doc:        existingDoc,
	}

	if err := c.storage.Append(record); err != nil {
		return fmt.Errorf("failed to persist update: %w", err)
	}

	return nil
}

// UpdateMany updates all documents matching the filter
// Returns the number of documents updated
func (c *Collection) UpdateMany(filter map[string]interface{}, update map[string]interface{}) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for id, doc := range c.documents {
		if MatchesFilter(doc, filter) {
			// Merge update into document
			for key, value := range update {
				doc[key] = value
			}
			doc["id"] = id

			// Persist to disk
			record := StorageRecord{
				Collection: c.name,
				ID:         id,
				Doc:        doc,
			}

			if err := c.storage.Append(record); err != nil {
				return count, fmt.Errorf("failed to persist update: %w", err)
			}

			count++
		}
	}

	return count, nil
}

// Delete removes a document from the collection
func (c *Collection) Delete(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if document exists
	if _, exists := c.documents[id]; !exists {
		return fmt.Errorf("document with id %s not found", id)
	}

	// Remove from memory
	delete(c.documents, id)

	// Persist deletion to disk (nil document indicates deletion)
	record := StorageRecord{
		Collection: c.name,
		ID:         id,
		Doc:        nil,
	}

	if err := c.storage.Append(record); err != nil {
		return fmt.Errorf("failed to persist deletion: %w", err)
	}

	return nil
}

// DeleteMany deletes all documents matching the filter
// Returns the number of documents deleted
func (c *Collection) DeleteMany(filter map[string]interface{}) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	idsToDelete := []string{}

	// Find all matching documents
	for id, doc := range c.documents {
		if MatchesFilter(doc, filter) {
			idsToDelete = append(idsToDelete, id)
		}
	}

	// Delete each document
	for _, id := range idsToDelete {
		delete(c.documents, id)

		// Persist deletion to disk
		record := StorageRecord{
			Collection: c.name,
			ID:         id,
			Doc:        nil,
		}

		if err := c.storage.Append(record); err != nil {
			return count, fmt.Errorf("failed to persist deletion: %w", err)
		}

		count++
	}

	return count, nil
}

// Count returns the number of documents in the collection
func (c *Collection) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.documents)
}

// CountWhere returns the number of documents matching the filter
func (c *Collection) CountWhere(filter map[string]interface{}) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(filter) == 0 {
		return len(c.documents)
	}

	count := 0
	for _, doc := range c.documents {
		if MatchesFilter(doc, filter) {
			count++
		}
	}

	return count
}
