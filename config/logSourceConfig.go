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

func NewLogSourceConfig(cfg *Config) LogSourceConfig {
	return LogSourceConfig{
		LogConnectionStats: !cfg.TrafficLogger.NoLogConnStats,
		LogRequestHeaders:  !cfg.TrafficLogger.NoLogReqHeaders,
		LogRequest:         !cfg.TrafficLogger.NoLogReqBody,
		LogResponseHeaders: !cfg.TrafficLogger.NoLogRespHeaders,
		LogResponse:        !cfg.TrafficLogger.NoLogRespBody,
	}
}

func (l *LogSourceConfig) String() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		slog.Error("Could not load LogSourceConfig", "error", err)
		return ""
	}
	return string(bytes)
}

var LogSourceConfigAllTrue = LogSourceConfig{
	LogConnectionStats: true,
	LogRequestHeaders:  true,
	LogRequest:         true,
	LogResponseHeaders: true,
	LogResponse:        true,
}
