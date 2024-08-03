package mitm

import (
	"maps"
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"
)

type ProxyResponseAdapter struct {
	pxResp     *px.Response
	headerCopy http.Header
}

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

func (r *ProxyResponseAdapter) GetStatusCode() int {
	return r.pxResp.StatusCode
}

func (r *ProxyResponseAdapter) GetHeaders() http.Header {
	return r.headerCopy
}

func (r *ProxyResponseAdapter) GetBodyBytes() []byte {
	return r.pxResp.Body
}
