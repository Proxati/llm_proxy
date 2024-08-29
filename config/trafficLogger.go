package config

import (
	"encoding/json"
	"log/slog"
)

// LogSourceConfig holds the configuration toggles for logging request and response data
type LogSourceConfig struct {
	LogConnectionStats bool
	LogRequestHeaders  bool
	LogRequest         bool
	LogResponseHeaders bool
	LogResponse        bool
}

func (l *LogSourceConfig) String() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		slog.Error("Could not load LogSourceConfig", "error", err)
		return ""
	}
	return string(bytes)
}

// LogSourceConfigAllTrue is a LogSourceConfig with all fields set to true
var LogSourceConfigAllTrue = LogSourceConfig{
	LogConnectionStats: true,
	LogRequestHeaders:  true,
	LogRequest:         true,
	LogResponseHeaders: true,
	LogResponse:        true,
}

// trafficLogger stores config filtering options for the *output* of the proxy traffic. To
// turn off logging of a part of the transaction, if needed.
type trafficLogger struct {
	Output           string    // Directory or Address to write logs
	LogFmt           LogFormat // Traffic log output format (json, txt)
	NoLogConnStats   bool      // if true, do not log connection stats
	NoLogReqHeaders  bool      // if true, log request headers
	NoLogReqBody     bool      // if true, log request body
	NoLogRespHeaders bool      // if true, log response headers
	NoLogRespBody    bool      // if true, log response body
}

func (t *trafficLogger) GetLogSourceConfig() LogSourceConfig {
	return LogSourceConfig{
		LogConnectionStats: !t.NoLogConnStats,
		LogRequestHeaders:  !t.NoLogReqHeaders,
		LogRequest:         !t.NoLogReqBody,
		LogResponseHeaders: !t.NoLogRespHeaders,
		LogResponse:        !t.NoLogRespBody,
	}
}
