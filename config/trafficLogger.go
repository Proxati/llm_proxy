package config

// trafficLogger handles config related to the *output* of the proxy traffic, for writing request/response logs
type trafficLogger struct {
	Output           string    // Directory or Address to write logs
	TrafficLogFmt    LogFormat // Traffic log output format (json, txt)
	NoLogConnStats   bool      // if true, do not log connection stats
	NoLogReqHeaders  bool      // if true, log request headers
	NoLogReqBody     bool      // if true, log request body
	NoLogRespHeaders bool      // if true, log response headers
	NoLogRespBody    bool      // if true, log response body
}
