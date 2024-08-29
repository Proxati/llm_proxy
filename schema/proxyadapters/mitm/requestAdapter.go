package mitm

import (
	"maps"
	"net/http"
	"net/url"

	px "github.com/proxati/mitmproxy/proxy"
)

// ProxyRequestAdapter is a RequestAdapter implementation for mitmproxy requests
type ProxyRequestAdapter struct {
	pxReq      *px.Request
	headerCopy http.Header
}

// NewProxyRequestAdapter creates a new request adapter object
func NewProxyRequestAdapter(pxReq *px.Request) *ProxyRequestAdapter {
	if pxReq == nil {
		return &ProxyRequestAdapter{
			pxReq:      &px.Request{Header: http.Header{}, URL: &url.URL{}},
			headerCopy: http.Header{},
		}
	}
	if pxReq.Header == nil {
		pxReq.Header = http.Header{}
	}
	if pxReq.URL == nil {
		pxReq.URL = &url.URL{}
	}
	// shallow copy of the headers to prevent race conditions
	headerCopy := http.Header{}
	maps.Copy(headerCopy, pxReq.Header)

	return &ProxyRequestAdapter{pxReq: pxReq, headerCopy: headerCopy}
}

// GetMethod returns the HTTP request method, to implement the RequestReaderAdapter interface
func (r *ProxyRequestAdapter) GetMethod() string {
	return r.pxReq.Method
}

// GetURL returns the request URL, to implement the RequestReaderAdapter interface
func (r *ProxyRequestAdapter) GetURL() *url.URL {
	return r.pxReq.URL
}

// GetProto returns the HTTP protocol version, to implement the RequestReaderAdapter interface
func (r *ProxyRequestAdapter) GetProto() string {
	return r.pxReq.Proto
}

// GetHeaders returns the headers, to implement the RequestReaderAdapter interface
func (r *ProxyRequestAdapter) GetHeaders() http.Header {
	return r.headerCopy
}

// GetBodyBytes returns the request body, to implement the RequestReaderAdapter interface
func (r *ProxyRequestAdapter) GetBodyBytes() []byte {
	return r.pxReq.Body
}
