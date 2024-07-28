package schema

import (
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"
)

type ProxyResponseReaderAdapter interface {
	GetStatusCode() int
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

type ProxyResponseAdapter_MiTM struct {
	pxResp *px.Response
}

func NewProxyResponseAdapter_MiTM(pxResp *px.Response) *ProxyResponseAdapter_MiTM {
	if pxResp == nil {
		return nil
	}

	return &ProxyResponseAdapter_MiTM{pxResp: pxResp}
}

func (r *ProxyResponseAdapter_MiTM) GetStatusCode() int {
	return r.pxResp.StatusCode
}

func (r *ProxyResponseAdapter_MiTM) GetHeaders() http.Header {
	return r.pxResp.Header
}

func (r *ProxyResponseAdapter_MiTM) GetBodyBytes() []byte {
	return r.pxResp.Body
}
