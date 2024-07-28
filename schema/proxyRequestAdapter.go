package schema

import (
	"net/http"
	"net/url"

	px "github.com/proxati/mitmproxy/proxy"
)

type ProxyRequestReaderAdapter interface {
	GetMethod() string
	GetURL() *url.URL
	GetProto() string
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

// ProxyRequestAdapter_MiTM is a RequestAdapter implementation for mitmproxy requests
type ProxyRequestAdapter_MiTM struct {
	pxReq *px.Request
}

func NewProxyRequestAdapter_MiTM(pxReq *px.Request) *ProxyRequestAdapter_MiTM {
	if pxReq == nil {
		return nil
	}

	return &ProxyRequestAdapter_MiTM{pxReq: pxReq}
}

func (r *ProxyRequestAdapter_MiTM) GetMethod() string {
	return r.pxReq.Method
}

func (r *ProxyRequestAdapter_MiTM) GetURL() *url.URL {
	return r.pxReq.URL
}

func (r *ProxyRequestAdapter_MiTM) GetProto() string {
	return r.pxReq.Proto
}

func (r *ProxyRequestAdapter_MiTM) GetHeaders() http.Header {
	return r.pxReq.Header
}

func (r *ProxyRequestAdapter_MiTM) GetBodyBytes() []byte {
	return r.pxReq.Body
}
