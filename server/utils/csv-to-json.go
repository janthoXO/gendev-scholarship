package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CSVToMap converts CSV data to a slice of maps where each map represents a row
// with column names as keys and cell values as values
func CSVToMap(csvData []byte) ([]map[string]interface{}, error) {
	reader := csv.NewReader(bytes.NewReader(csvData))

	// Read header row to get column names
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Trim whitespace from headers
	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}

	// Parse the rest of the CSV into a slice of maps
	var result []map[string]interface{}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})

		// Map each column value to its header and convert to appropriate type
		for i, value := range record {
			if i < len(headers) {
				trimmedValue := strings.TrimSpace(value)
				row[headers[i]] = convertStringValue(trimmedValue)
			}
		}

		result = append(result, row)
	}

	return result, nil
}

// convertStringValue attempts to convert a string value to the appropriate type
// Useful for type conversion when populating structs
func convertStringValue(value string) interface{} {
	// Try to convert to int
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Try to convert to float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Check for boolean values
	switch strings.ToLower(value) {
	case "true", "yes", "1":
		return true
	case "false", "no", "0":
		return false
	}

	// Return as string if no conversion was successful
	return value
}
