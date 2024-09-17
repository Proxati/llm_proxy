package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
)

// SchemaVersion is the version of the schema, used for backwards compatibility.
const (
	SchemaVersionV2      string = "v2"
	DefaultSchemaVersion string = SchemaVersionV2
)

// ObjectTypeDefault is the default object type for the log dump container, used for identifying
// the file type when loading various objects from json files.
const ObjectTypeDefault string = "llm_proxy_traffic_log"

// LogDumpContainer holds the request and response data for a given flow
type LogDumpContainer struct {
	ObjectType      string                `json:"object_type,omitempty"`
	SchemaVersion   string                `json:"schema,omitempty"`
	Timestamp       time.Time             `json:"timestamp,omitempty"`
	ConnectionStats *ProxyConnectionStats `json:"connection_stats,omitempty"`
	Request         *ProxyRequest         `json:"request,omitempty"`
	Response        *ProxyResponse        `json:"response,omitempty"`
	logConfig       config.LogSourceConfig
}

func NewLogDumpContainerEmpty() *LogDumpContainer {
	return &LogDumpContainer{
		ObjectType:    ObjectTypeDefault,
		SchemaVersion: DefaultSchemaVersion,
		Timestamp:     time.Now(),
		Request:       &ProxyRequest{URL: &url.URL{}, Header: make(http.Header)},
		Response:      &ProxyResponse{Header: make(http.Header)},
	}
}

// NewLogDumpContainerFromFlowAdapter returns a LogDumpContainer with *only* the fields requested in logSources populated
func NewLogDumpContainerFromFlowAdapter(
	f proxyadapters.FlowReaderAdapter,
	logSources config.LogSourceConfig,
	doneAt int64,
	filterReqHeaders, filterRespHeaders *config.HeaderFilterGroup,
) (*LogDumpContainer, error) {
	if f == nil {
		return nil, errors.New("flow is nil")
	}

	var err error
	errs := make([]error, 0)
	ldc := NewLogDumpContainerEmpty()
	ldc.logConfig = logSources

	if logSources.LogRequest {
		// convert the request to a request adapter
		ldc.Request, err = NewProxyRequest(f.GetRequest(), filterReqHeaders)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if !logSources.LogRequestHeaders {
		ldc.Request.Header = nil
	}

	if logSources.LogResponse {
		ldc.Response, err = NewProxyResponse(f.GetResponse(), filterRespHeaders)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if !logSources.LogResponseHeaders {
		ldc.Response.Header = nil
	}

	if logSources.LogConnectionStats {
		ldc.ConnectionStats = NewProxyConnectionStatsWithDuration(f.GetConnectionStats(), doneAt)
	}

	if len(errs) > 0 {
		for _, err := range errs {
			if err != nil {
				// TODO: need to reconsider how to handle errors here
				getLogger().Error("errors encountered while creating LogDumpContainer", "error", err)
			}
		}
		return ldc, fmt.Errorf("errors encountered while creating LogDumpContainer: %w", errors.Join(errs...))
	}

	return ldc, nil
}

func (ldc *LogDumpContainer) unmarshalJSON_V2(data []byte) error {
	// Create a temporary struct to unmarshal the top-level fields
	type Alias LogDumpContainer
	aux := &struct {
		*Alias
		Timestamp       string          `json:"timestamp,omitempty"`
		ConnectionStats json.RawMessage `json:"connection_stats,omitempty"`
		Request         json.RawMessage `json:"request,omitempty"`
		Response        json.RawMessage `json:"response,omitempty"`
	}{
		Alias: (*Alias)(ldc),
	}

	// Unmarshal the raw JSON into the temporary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse the timestamp if provided
	if aux.Timestamp != "" {
		t, err := time.Parse(time.RFC3339, aux.Timestamp)
		if err != nil {
			return err
		}
		ldc.Timestamp = t
	}

	// Unmarshal connection stats if they exist
	if aux.ConnectionStats != nil {
		ldc.ConnectionStats = &ProxyConnectionStats{}
		if err := json.Unmarshal(aux.ConnectionStats, ldc.ConnectionStats); err != nil {
			return err
		}
	}

	// Unmarshal request if it exists
	if aux.Request != nil {
		ldc.Request = &ProxyRequest{}
		if err := json.Unmarshal(aux.Request, ldc.Request); err != nil {
			return err
		}
	}

	// Unmarshal response if it exists
	if aux.Response != nil {
		ldc.Response = &ProxyResponse{}
		if err := json.Unmarshal(aux.Response, ldc.Response); err != nil {
			return err
		}
	}

	ldc.SchemaVersion = SchemaVersionV2
	return nil
}

func (ldc *LogDumpContainer) UnmarshalJSON(data []byte) error {
	r := make(map[string]interface{})
	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	// filter unknown/unset object_type
	objectType, ok := r["object_type"]
	if !ok {
		return errors.New("object_type is required")
	}

	objectTypeStr, ok := objectType.(string)
	if !ok {
		return errors.New("object_type must be a string")
	}

	if objectTypeStr != ObjectTypeDefault {
		return errors.New("unsupported object_type: " + objectTypeStr)
	}

	// handle schema version
	schemaVersion, ok := r["schema"]
	if !ok {
		// assume latest schema version if not provided
		schemaVersion = DefaultSchemaVersion
	}

	schemaVersionStr, ok := schemaVersion.(string)
	if !ok {
		return errors.New("schema must be a string")
	}

	switch schemaVersionStr {
	case SchemaVersionV2:
		return ldc.unmarshalJSON_V2(data)
	default:
		return errors.New("unsupported schema version")
	}
}

// UnmarshalLogDumpContainer unmarshals JSON data into a LogDumpContainer
func UnmarshalLogDumpContainer(data []byte) (*LogDumpContainer, error) {
	var ldc LogDumpContainer
	err := json.Unmarshal(data, &ldc)
	if err != nil {
		return nil, err
	}
	return &ldc, nil
}

func (ldc *LogDumpContainer) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid infinite recursion
	type Alias LogDumpContainer

	// Format the timestamp if it's not zero
	var timestamp string
	if !ldc.Timestamp.IsZero() {
		timestamp = ldc.Timestamp.Format(time.RFC3339)
	}

	// Construct the temporary struct for marshaling
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp,omitempty"`
	}{
		Alias:     (*Alias)(ldc),
		Timestamp: timestamp,
	}

	// Marshal the temporary struct to JSON
	return json.Marshal(aux)
}
