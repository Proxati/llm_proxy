package mitm

import (
	"net/http"
	"net/url"

	px "github.com/proxati/mitmproxy/proxy"
)

// ProxyRequestAdapter is a RequestAdapter implementation for mitmproxy requests
type ProxyRequestAdapter struct {
	pxReq *px.Request
}

func NewProxyRequestAdapter(pxReq *px.Request) *ProxyRequestAdapter {
	if pxReq == nil {
		return nil
	}

	return &ProxyRequestAdapter{pxReq: pxReq}
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
	return r.pxReq.Header
}

func (r *ProxyRequestAdapter) GetBodyBytes() []byte {
	return r.pxReq.Body
}
