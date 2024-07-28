package schema

import (
	"net/http"
	"net/url"

	px "github.com/proxati/mitmproxy/proxy"
)

type RequestAdapter interface {
	GetMethod() string
	GetURL() *url.URL
	GetProto() string
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

// RequestAdapter_MiTM is a RequestAccessor implementation for mitmproxy requests
type RequestAdapter_MiTM struct {
	pxReq *px.Request
}

func NewRequestAdapter_MiTM(pxReq *px.Request) *RequestAdapter_MiTM {
	if pxReq == nil {
		return nil
	}

	return &RequestAdapter_MiTM{pxReq: pxReq}
}

func (r *RequestAdapter_MiTM) GetMethod() string {
	return r.pxReq.Method
}

func (r *RequestAdapter_MiTM) GetURL() *url.URL {
	return r.pxReq.URL
}

func (r *RequestAdapter_MiTM) GetProto() string {
	return r.pxReq.Proto
}

func (r *RequestAdapter_MiTM) GetHeaders() http.Header {
	return r.pxReq.Header
}

func (r *RequestAdapter_MiTM) GetBodyBytes() []byte {
	return r.pxReq.Body
}
