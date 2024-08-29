package config

import (
	"fmt"
	"strings"
)

// LogFormat is an enum that represents different log formats
type LogFormat int

const (
	// LogFormatJSON is the JSON log format
	LogFormatJSON LogFormat = iota

	// LogFormatTXT is the plain text log format
	LogFormatTXT
)

func (f LogFormat) String() string {
	switch f {
	case LogFormatJSON:
		return "json"
	case LogFormatTXT:
		return "txt"
	default:
		return ""
	}
}

// StringToLogFormat converts a string to a LogLevel enum value.
func StringToLogFormat(logFormat string) (LogFormat, error) {
	logFormat = strings.TrimSpace(logFormat)
	logFormat = strings.ToLower(logFormat)

	switch logFormat {
	case "json":
		return LogFormatJSON, nil
	case "txt", "text":
		return LogFormatTXT, nil
	default:
		return 0, fmt.Errorf("log format not supported: %s", logFormat)
	}
}
