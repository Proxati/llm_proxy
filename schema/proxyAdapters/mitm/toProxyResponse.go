package mitm

import (
	"fmt"

	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	px "github.com/proxati/mitmproxy/proxy"
)

// ToProxyResponse converts a ProxyResponse into a MITM proxy response object (with content encoding matching the new req)
// Because all responses are stored as uncompressed strings, the cached response might need to be encoded before being sent
func ToProxyResponse(pRes proxyAdapters.ResponseReaderAdapter, acceptEncodingHeader string) (*px.Response, error) {
	resp := &px.Response{
		StatusCode: pRes.GetStatusCode(),
		Header:     pRes.GetHeaders(),
	}
	body := pRes.GetBodyBytes()

	// encode the body based on the new request's accept encoding header
	encodedBody, encoding, err := utils.EncodeBody(body, acceptEncodingHeader)
	if err != nil {
		return nil, fmt.Errorf("error encoding body: %v", err)
	}

	// set the new content encoding and length headers
	resp.Header.Set("Content-Encoding", encoding)
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

	resp.Body = encodedBody
	return resp, nil
}
