package mitm

import (
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"
)

type ProxyResponseAdapter struct {
	pxResp *px.Response
}

func NewProxyResponseAdapter(pxResp *px.Response) *ProxyResponseAdapter {
	if pxResp == nil {
		return nil
	}

	return &ProxyResponseAdapter{pxResp: pxResp}
}

func (r *ProxyResponseAdapter) GetStatusCode() int {
	return r.pxResp.StatusCode
}

func (r *ProxyResponseAdapter) GetHeaders() http.Header {
	return r.pxResp.Header
}

func (r *ProxyResponseAdapter) GetBodyBytes() []byte {
	return r.pxResp.Body
}
