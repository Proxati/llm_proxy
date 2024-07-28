package schema

import (
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"
)

type ResponseAdapter interface {
	GetStatusCode() int
	GetHeaders() http.Header
	GetBodyBytes() []byte
}

type ResponseAdapter_MiTM struct {
	pxResp *px.Response
}

func NewResponseAdapter_MiTM(pxResp *px.Response) *ResponseAdapter_MiTM {
	if pxResp == nil {
		return nil
	}

	return &ResponseAdapter_MiTM{pxResp: pxResp}
}

func (r *ResponseAdapter_MiTM) GetStatusCode() int {
	return r.pxResp.StatusCode
}

func (r *ResponseAdapter_MiTM) GetHeaders() http.Header {
	return r.pxResp.Header
}

func (r *ResponseAdapter_MiTM) GetBodyBytes() []byte {
	return r.pxResp.Body
}
