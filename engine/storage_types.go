package engine

// StorageRecord represents a single record in the storage
// Each record is a JSON-encoded StorageRecord
type StorageRecord struct {
	Collection string                 `json:"collection"` // Name of the collection
	ID         string                 `json:"id"`         // Unique document ID
	Doc        map[string]interface{} `json:"doc"`        // The actual document data
}
