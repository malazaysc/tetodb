// +build js,wasm

package engine

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Storage handles in-memory persistence for WASM builds
// Since file I/O is not available in GOOS=js, we keep everything in memory
type Storage struct {
	filePath string             // Logical path (for compatibility)
	records  []StorageRecord    // All records in memory
	mu       sync.Mutex         // Protects concurrent access
}

// NewStorage creates a new in-memory Storage instance
func NewStorage(path string) (*Storage, error) {
	return &Storage{
		filePath: path,
		records:  make([]StorageRecord, 0),
	}, nil
}

// LoadAll returns all records from memory
func (s *Storage) LoadAll() ([]StorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy to avoid external modification
	result := make([]StorageRecord, len(s.records))
	copy(result, s.records)
	return result, nil
}

// Append adds a new record to memory
func (s *Storage) Append(record StorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate that the record can be serialized (for consistency with file-based version)
	if _, err := json.Marshal(record); err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	s.records = append(s.records, record)
	return nil
}

// Close is a no-op for in-memory storage
func (s *Storage) Close() error {
	return nil
}

// Compact rebuilds the in-memory records
func (s *Storage) Compact(records []StorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Replace all records with the compacted set
	s.records = make([]StorageRecord, len(records))
	copy(s.records, records)

	return nil
}
