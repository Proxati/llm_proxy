package proxyAdapters

import (
	"net/http"
	"net/url"
)

type RequestReaderAdapter interface {
	GetMethod() string
	GetURL() *url.URL
	GetProto() string
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

type ResponseReaderAdapter interface {
	GetStatusCode() int
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

type ConnectionStatsReaderAdapter interface {
	GetClientIP() string
	GetProxyID() string
	GetRequestURL() string
}

type FlowReaderAdapter interface {
	GetRequest() RequestReaderAdapter
	GetResponse() ResponseReaderAdapter
	GetConnectionStats() ConnectionStatsReaderAdapter
}
