package formatters

import (
	"encoding/json"
	"fmt"

	"github.com/proxati/llm_proxy/schema"
)

type JSON struct {
	container *schema.LogDumpContainer
}

// dumpToJSONBytes converts the requestLogDump struct to a byte array, omitting fields that are empty
func (f *JSON) dumpToJSONBytes() ([]byte, error) {
	if f.container == nil {
		return []byte("{}"), nil
	}

	j, err := json.MarshalIndent(f.container, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal requestLogDump to JSON: %w", err)
	}
	return j, nil
}

// Read returns the JSON representation of a LogDumpContainer (json formatted byte array)
func (f *JSON) Read(container *schema.LogDumpContainer) ([]byte, error) {
	f.container = container
	return f.dumpToJSONBytes()
}
