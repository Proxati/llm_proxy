package proxyadapters

import (
	"net/http"
	"net/url"
)

// RequestReaderAdapter is an interface for reading request data from any proxy object that has
// been abstracted by this interface.
type RequestReaderAdapter interface {
	GetMethod() string
	GetURL() *url.URL
	GetProto() string
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

// ResponseReaderAdapter is an interface for reading response data from any proxy object that has
// been abstracted by this interface.
type ResponseReaderAdapter interface {
	GetStatusCode() int
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

// ConnectionStatsReaderAdapter is an interface for reading connection stats data from any proxy
// connection stats object that has been abstracted by this interface. Connection stats are
// information about the connection between the client and the proxy.
type ConnectionStatsReaderAdapter interface {
	GetClientIP() string
	GetProxyID() string
	GetRequestURL() string
}

// FlowReaderAdapter is an interface for reading flow data from any proxy flow object that has
// been abstracted by this interface. A flow is a collection of data about a single request,
// response, and connection stats.
//
// This interface is the first step in replacing the mitmproxy-specific flow object with a more
// generic flow object that can be used with any proxy library.
type FlowReaderAdapter interface {
	GetRequest() RequestReaderAdapter
	GetResponse() ResponseReaderAdapter
	GetConnectionStats() ConnectionStatsReaderAdapter
}
