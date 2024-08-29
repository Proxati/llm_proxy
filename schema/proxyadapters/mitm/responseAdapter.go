package mitm

import (
	"maps"
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"
)

// ProxyResponseAdapter implements the proxyAdapters.ResponseReaderAdapter interface
type ProxyResponseAdapter struct {
	pxResp     *px.Response
	headerCopy http.Header
}

// NewProxyResponseAdapter creates a new response adapter object
func NewProxyResponseAdapter(pxResp *px.Response) *ProxyResponseAdapter {
	if pxResp == nil {
		return &ProxyResponseAdapter{
			pxResp:     &px.Response{Header: http.Header{}},
			headerCopy: http.Header{},
		}
	}
	if pxResp.Header == nil {
		pxResp.Header = http.Header{}
	}

	// shallow copy of the headers to prevent race conditions
	headerCopy := http.Header{}
	maps.Copy(headerCopy, pxResp.Header)

	return &ProxyResponseAdapter{pxResp: pxResp, headerCopy: headerCopy}
}

// GetStatusCode returns the status code, to implement the ResponseReaderAdapter interface
func (r *ProxyResponseAdapter) GetStatusCode() int {
	return r.pxResp.StatusCode
}

// GetHeaders returns the headers, to implement the ResponseReaderAdapter interface
func (r *ProxyResponseAdapter) GetHeaders() http.Header {
	return r.headerCopy
}

// GetBodyBytes returns the response body, to implement the ResponseReaderAdapter interface
func (r *ProxyResponseAdapter) GetBodyBytes() []byte {
	return r.pxResp.Body
}
