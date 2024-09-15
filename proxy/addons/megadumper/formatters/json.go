package formatters

import (
	"encoding/json"
	"fmt"

	"github.com/proxati/llm_proxy/v2/schema"
)

const jsonExt = "json"

// JSON is a formatter that converts the LogDumpContainer into a JSON formatted byte array
type JSON struct{}

// Read returns the JSON representation of a LogDumpContainer (JSON formatted byte array)
func (f *JSON) Read(container *schema.LogDumpContainer) ([]byte, error) {
	return f.dumpToJSONBytes(container)
}

// dumpToJSONBytes converts the LogDumpContainer struct to a byte array
func (f *JSON) dumpToJSONBytes(container *schema.LogDumpContainer) ([]byte, error) {
	if container == nil {
		return []byte("{}"), nil
	}

	j, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LogDumpContainer to JSON: %w", err)
	}
	return j, nil
}

// GetFileExtension returns the file extension for a JSON file
func (f *JSON) GetFileExtension() string {
	return jsonExt
}

// String returns the name of the formatter
func (f *JSON) String() string {
	return "JSON"
}
