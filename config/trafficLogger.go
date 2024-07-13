package config

import (
	"fmt"
	"strings"
)

type TrafficLogFormat int

const (
	TrafficLog_JSON TrafficLogFormat = iota
	TrafficLog_TXT
)

// StringToTrafficLogFormat converts a string to a LogLevel enum value.
func StringToTrafficLogFormat(logFormat string) (TrafficLogFormat, error) {
	switch strings.ToLower(logFormat) {
	case "json":
		return TrafficLog_JSON, nil
	case "txt", "text":
		return TrafficLog_TXT, nil
	default:
		return 0, fmt.Errorf("log format not supported: %s", logFormat)
	}
}

// trafficLogger handles config related to the *output* of the proxy traffic, for writing request/response logs
type trafficLogger struct {
	OutputDir         string           // Directory to write logs
	LogFormat         TrafficLogFormat // Traffic log output format (json, txt)
	NoLogConnStats    bool             // if true, do not log connection stats
	NoLogReqHeaders   bool             // if true, log request headers
	NoLogReqBody      bool             // if true, log request body
	NoLogRespHeaders  bool             // if true, log response headers
	NoLogRespBody     bool             // if true, log response body
	FilterReqHeaders  []string         // if set, request headers that match these strings will not be logged
	FilterRespHeaders []string         // if set, response headers that match these strings will not be logged
}
