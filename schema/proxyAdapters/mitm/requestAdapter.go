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

func (r *ProxyRequestAdapter) GetMethod() string {
	return r.pxReq.Method
}

func (r *ProxyRequestAdapter) GetURL() *url.URL {
	return r.pxReq.URL
}

func (r *ProxyRequestAdapter) GetProto() string {
	return r.pxReq.Proto
}

func (r *ProxyRequestAdapter) GetHeaders() http.Header {
	return r.headerCopy
}

func (r *ProxyRequestAdapter) GetBodyBytes() []byte {
	return r.pxReq.Body
}
