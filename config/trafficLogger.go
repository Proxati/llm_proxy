package config

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
