package engine

import (
	"fmt"
	"reflect"
	"strings"
)

// MatchesFilter checks if a document matches the given filter
// Supports basic equality matching and simple operators
//
// Filter format examples:
//   {"name": "John"}                    // Exact match
//   {"age": 25}                         // Numeric match
//   {"status": "active", "role": "admin"} // AND condition (all must match)
//
// Note: This is a simple implementation for demonstration purposes
// A production system would support more complex queries ($gt, $lt, $in, etc.)
func MatchesFilter(doc map[string]interface{}, filter map[string]interface{}) bool {
	// Empty filter matches everything
	if len(filter) == 0 {
		return true
	}

	// All filter conditions must match (AND logic)
	for key, filterValue := range filter {
		docValue, exists := doc[key]

		// If field doesn't exist in document, no match
		if !exists {
			return false
		}

		// Check if values match
		if !valuesMatch(docValue, filterValue) {
			return false
		}
	}

	return true
}

// valuesMatch compares two values for equality
// Handles different types and attempts to do smart comparison
func valuesMatch(docValue, filterValue interface{}) bool {
	// Direct equality check
	if reflect.DeepEqual(docValue, filterValue) {
		return true
	}

	// Convert both to strings for comparison if types differ
	docStr := fmt.Sprintf("%v", docValue)
	filterStr := fmt.Sprintf("%v", filterValue)

	return docStr == filterStr
}

// QueryBuilder provides a fluent API for building queries (future enhancement)
// This is a placeholder for more advanced query functionality
type QueryBuilder struct {
	filter map[string]interface{}
}

// NewQuery creates a new QueryBuilder
func NewQuery() *QueryBuilder {
	return &QueryBuilder{
		filter: make(map[string]interface{}),
	}
}

// Where adds an equality condition to the query
func (q *QueryBuilder) Where(field string, value interface{}) *QueryBuilder {
	q.filter[field] = value
	return q
}

// Build returns the constructed filter
func (q *QueryBuilder) Build() map[string]interface{} {
	return q.filter
}

// ParseFilterString parses a simple filter string into a filter map
// Format: "field1=value1,field2=value2"
// This is useful for converting string-based queries from the WASM layer
func ParseFilterString(filterStr string) map[string]interface{} {
	filter := make(map[string]interface{})

	if filterStr == "" {
		return filter
	}

	// Split by comma to get individual conditions
	parts := strings.Split(filterStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by = to get field and value
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		field := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Store in filter map
		filter[field] = value
	}

	return filter
}

// SortDocuments sorts documents by a field (simple implementation)
// direction: "asc" or "desc"
// Note: This modifies the slice in place
func SortDocuments(docs []map[string]interface{}, field string, direction string) {
	// For simplicity, we'll use a basic bubble sort
	// A production system would use a more efficient algorithm
	n := len(docs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			val1, exists1 := docs[j][field]
			val2, exists2 := docs[j+1][field]

			// Handle missing fields
			if !exists1 || !exists2 {
				continue
			}

			// Compare values
			shouldSwap := false
			if direction == "desc" {
				shouldSwap = compareValues(val1, val2) < 0
			} else {
				shouldSwap = compareValues(val1, val2) > 0
			}

			if shouldSwap {
				docs[j], docs[j+1] = docs[j+1], docs[j]
			}
		}
	}
}

// compareValues compares two values and returns:
//   -1 if a < b
//    0 if a == b
//    1 if a > b
func compareValues(a, b interface{}) int {
	// Try numeric comparison first
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if aOk && bOk {
		if aFloat < bFloat {
			return -1
		} else if aFloat > bFloat {
			return 1
		}
		return 0
	}

	// Fall back to string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// toFloat64 attempts to convert a value to float64
func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}
