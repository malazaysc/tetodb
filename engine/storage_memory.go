// +build js,wasm

package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"syscall/js"
)

// Storage handles file-based persistence for WASM builds
// Uses Node.js file system operations via JavaScript bridge
type Storage struct {
	filePath string     // Path to the database file
	mu       sync.Mutex // Protects concurrent access
}

// NewStorage creates a new Storage instance and ensures file exists
func NewStorage(path string) (*Storage, error) {
	s := &Storage{
		filePath: path,
	}

	// Try to read the file to ensure it exists (will create if not)
	nodeFileRead := js.Global().Get("nodeFileReadSync")
	if !nodeFileRead.Truthy() {
		return nil, fmt.Errorf("nodeFileReadSync not available - Node.js file system bridge not initialized")
	}

	// Read file (creates it if it doesn't exist by returning empty string)
	_ = nodeFileRead.Invoke(path)

	return s, nil
}

// LoadAll reads all records from the file
func (s *Storage) LoadAll() ([]StorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeFileRead := js.Global().Get("nodeFileReadSync")
	if !nodeFileRead.Truthy() {
		return nil, fmt.Errorf("nodeFileReadSync not available")
	}

	// Read entire file
	result := nodeFileRead.Invoke(s.filePath)
	content := result.String()

	if content == "" {
		return []StorageRecord{}, nil
	}

	// Parse line by line
	var records []StorageRecord
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var record StorageRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			// Log error but continue
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

// Append writes a new record to the file
func (s *Storage) Append(record StorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeFileAppend := js.Global().Get("nodeFileAppendSync")
	if !nodeFileAppend.Truthy() {
		return fmt.Errorf("nodeFileAppendSync not available")
	}

	// Serialize record to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// Append with newline
	line := string(data) + "\n"
	nodeFileAppend.Invoke(s.filePath, line)

	return nil
}

// Close is a no-op for WASM storage (file operations are synchronous)
func (s *Storage) Close() error {
	return nil
}

// Compact rebuilds the file with only current records
func (s *Storage) Compact(records []StorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeFileWrite := js.Global().Get("nodeFileWriteSync")
	nodeFileRename := js.Global().Get("nodeFileRenameSync")
	nodeFileUnlink := js.Global().Get("nodeFileUnlinkSync")

	if !nodeFileWrite.Truthy() || !nodeFileRename.Truthy() || !nodeFileUnlink.Truthy() {
		return fmt.Errorf("Node.js file operations not available")
	}

	// Create temp file path
	tempPath := s.filePath + ".tmp"

	// Build content for temp file
	var content strings.Builder
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record: %w", err)
		}
		content.Write(data)
		content.WriteString("\n")
	}

	// Write to temp file
	nodeFileWrite.Invoke(tempPath, content.String())

	// Delete old file if it exists
	nodeFileUnlink.Invoke(s.filePath)

	// Rename temp file to actual file
	nodeFileRename.Invoke(tempPath, s.filePath)

	return nil
}
