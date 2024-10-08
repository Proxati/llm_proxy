package mitm

import (
	"fmt"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	px "github.com/proxati/mitmproxy/proxy"
)

// ToProxyResponse converts a ProxyResponse into a MITM proxy response object with body encoding
// matching the new request's acceptEncodingHeader. Since all responses are stored as uncompressed
// strings, the cached response might need to be encoded before being sent. This function encodes
// the response body based on the requested acceptEncodingHeader argument.
func ToProxyResponse(pRes proxyadapters.ResponseReaderAdapter, acceptEncodingHeader string) (*px.Response, error) {
	resp := &px.Response{
		StatusCode: pRes.GetStatusCode(),
		Header:     pRes.GetHeaders(),
	}
	body := pRes.GetBodyBytes()

	// encode the body based on the new request's accept encoding header
	encodedBody, encoding, err := utils.EncodeBody(body, acceptEncodingHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to encode response body with encoding '%s': %w", acceptEncodingHeader, err)
	}

	// set the new content encoding and length headers
	if encoding != "" {
		resp.Header.Set("Content-Encoding", encoding)
	} else {
		resp.Header.Del("Content-Encoding")
	}
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

	resp.Body = encodedBody
	return resp, nil
}
