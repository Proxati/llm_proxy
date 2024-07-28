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

/*
// ToProxyResponse converts a ProxyResponse into a MITM proxy response object (with content encoding matching the new req)
// Because all responses are stored as uncompressed strings, the cached response might need to be encoded before being sent
// TODO: move this out of the base schema package!
func (pRes *ProxyResponse) ToProxyResponse(acceptEncodingHeader string) (*px.Response, error) {
	resp := &px.Response{
		StatusCode: pRes.Status,
		Header:     pRes.Header,
	}

	// encode the body based on the new request's accept encoding header
	encodedBody, encoding, err := utils.EncodeBody(&pRes.Body, acceptEncodingHeader)
	if err != nil {
		return nil, fmt.Errorf("error encoding body: %v", err)
	}

	// set the new content encoding and length headers
	resp.Header.Set("Content-Encoding", encoding)
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

	resp.Body = encodedBody
	return resp, nil
}

*/
