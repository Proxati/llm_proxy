package schema

import (
	"errors"
	"net/url"
	"time"

	px "github.com/proxati/mitmproxy/proxy"

	"github.com/proxati/llm_proxy/v2/config"
)

const SchemaVersion string = "v2"
const ObjectTypeDefault string = "llm_proxy_traffic_log"

// LogDumpContainer holds the request and response data for a given flow
type LogDumpContainer struct {
	ObjectType      string                    `json:"object_type,omitempty"`
	SchemaVersion   string                    `json:"schema,omitempty"`
	Timestamp       time.Time                 `json:"timestamp,omitempty"`
	ConnectionStats *ConnectionStatsContainer `json:"connection_stats,omitempty"`
	Request         *ProxyRequest             `json:"request,omitempty"`
	Response        *ProxyResponse            `json:"response,omitempty"`
	logConfig       config.LogSourceConfig
}

// NewLogDumpContainer returns a LogDumpContainer with *only* the fields requested in logSources populated
func NewLogDumpContainer(f *px.Flow, logSources config.LogSourceConfig, doneAt int64, filterReqHeaders, filterRespHeaders []string) (*LogDumpContainer, error) {
	if f == nil {
		return nil, errors.New("flow is nil")
	}

	var err error
	errs := make([]error, 0)

	ldc := &LogDumpContainer{
		ObjectType:    ObjectTypeDefault,
		SchemaVersion: SchemaVersion,
		Timestamp:     time.Now(),
		logConfig:     logSources,
		Request: &ProxyRequest{
			URL: &url.URL{}, // NPE defense
		},
		Response: &ProxyResponse{},
	}

	if logSources.LogRequest {
		// convert the request to a request accessor
		reqAccessor, err := NewRequestAccessor(f.Request)
		if err != nil {
			return nil, err
		}

		ldc.Request, err = NewProxyRequest(reqAccessor, filterReqHeaders)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if !logSources.LogRequestHeaders {
		ldc.Request.Header = nil
	}

	if logSources.LogResponse {
		ldc.Response, err = NewProxyResponseFromMITMResponse(f.Response, filterRespHeaders)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if !logSources.LogResponseHeaders {
		ldc.Response.Header = nil
	}

	if logSources.LogConnectionStats {
		ldc.ConnectionStats = NewConnectionStatusContainerWithDuration(f, doneAt)
	}

	for _, err := range errs {
		if err != nil {
			// TODO: need to reconsider how to handle errors here
			getLogger().Error("errors encountered while creating LogDumpContainer", "error", err)
		}
	}

	return ldc, nil
}
