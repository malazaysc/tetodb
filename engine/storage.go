package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// StorageRecord represents a single record in the storage file
// Each line in the file is a JSON-encoded StorageRecord
type StorageRecord struct {
	Collection string                 `json:"collection"` // Name of the collection
	ID         string                 `json:"id"`         // Unique document ID
	Doc        map[string]interface{} `json:"doc"`        // The actual document data
}

// Storage handles the file-based persistence layer
// It uses a simple append-only log format where each line is a JSON record
type Storage struct {
	filePath string      // Path to the database file
	file     *os.File    // Open file handle
	mu       sync.Mutex  // Protects concurrent access to the file
}

// NewStorage creates a new Storage instance
// It opens (or creates) the file at the given path
func NewStorage(path string) (*Storage, error) {
	// Open file in read-write mode, create if doesn't exist
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage file: %w", err)
	}

	return &Storage{
		filePath: path,
		file:     file,
	}, nil
}

// LoadAll reads all records from the storage file
// Returns a slice of StorageRecords
func (s *Storage) LoadAll() ([]StorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Seek to beginning of file
	if _, err := s.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to beginning: %w", err)
	}

	var records []StorageRecord
	scanner := bufio.NewScanner(s.file)

	// Read line by line
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // Skip empty lines
		}

		var record StorageRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			// Log error but continue - don't let one corrupt record break everything
			fmt.Printf("Warning: failed to parse record: %v\n", err)
			continue
		}

		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return records, nil
}

// Append writes a new record to the end of the storage file
// Each record is written as a single JSON line
func (s *Storage) Append(record StorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Serialize record to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// Append newline-delimited JSON
	data = append(data, '\n')

	// Write to file
	if _, err := s.file.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	// Ensure data is flushed to disk
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// Close closes the storage file
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// Compact rebuilds the storage file by removing deleted/updated records
// This helps reclaim disk space from the append-only log
func (s *Storage) Compact(records []StorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close current file
	if err := s.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Create a temporary file
	tempPath := s.filePath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write all current records to temp file
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return fmt.Errorf("failed to marshal record: %w", err)
		}
		data = append(data, '\n')
		if _, err := tempFile.Write(data); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Replace old file with new file
	if err := os.Rename(tempPath, s.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Reopen the file
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen file: %w", err)
	}

	s.file = file
	return nil
}
