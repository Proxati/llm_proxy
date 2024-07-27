package schema

import (
	"errors"
	"net/http"
	"net/url"

	px "github.com/proxati/mitmproxy/proxy"
)

type RequestAccessor interface {
	GetMethod() string
	GetURL() *url.URL
	GetProto() string
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

// RequestAccessorMiTM is a RequestAccessor implementation for mitmproxy requests
type RequestAccessorMiTM struct {
	pxReq *px.Request
}

func NewRequestAccessor_MiTM(pxReq *px.Request) *RequestAccessorMiTM {
	return &RequestAccessorMiTM{pxReq: pxReq}
}

func (r *RequestAccessorMiTM) GetMethod() string {
	return r.pxReq.Method
}

func (r *RequestAccessorMiTM) GetURL() *url.URL {
	return r.pxReq.URL
}

func (r *RequestAccessorMiTM) GetProto() string {
	return r.pxReq.Proto
}

func (r *RequestAccessorMiTM) GetHeaders() http.Header {
	return r.pxReq.Header
}

func (r *RequestAccessorMiTM) GetBodyBytes() []byte {
	return r.pxReq.Body
}

// NewRequestAccessor returns a RequestAccessor for the given request object
func NewRequestAccessor(req any) (RequestAccessor, error) {
	switch req.(type) {
	case *px.Request:
		return NewRequestAccessor_MiTM(req.(*px.Request)), nil
	default:
		return nil, errors.New("unsupported request type")
	}
}
