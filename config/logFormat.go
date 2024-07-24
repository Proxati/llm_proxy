package config

import (
	"fmt"
	"strings"
)

type LogFormat int

const (
	LogFormat_JSON LogFormat = iota
	LogFormat_TXT
)

func (f LogFormat) String() string {
	switch f {
	case LogFormat_JSON:
		return "json"
	case LogFormat_TXT:
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
		return LogFormat_JSON, nil
	case "txt", "text":
		return LogFormat_TXT, nil
	default:
		return 0, fmt.Errorf("log format not supported: %s", logFormat)
	}
}
