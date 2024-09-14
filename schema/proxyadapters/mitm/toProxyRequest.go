package mitm

import (
	"fmt"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	px "github.com/proxati/mitmproxy/proxy"
)

// ToProxyRequest converts a RequestReaderAdapter into a MITM proxy request object with body encoding
// matching the new request's acceptEncodingHeader. Since all responses are stored as uncompressed
// strings, the cached response might need to be encoded before being sent. This function encodes
// the response body based on the requested acceptEncodingHeader argument and sets the appropriate
// headers for content encoding and content length.
func ToProxyRequest(pReq proxyadapters.RequestReaderAdapter, acceptEncodingHeader string) (*px.Request, error) {
	req := &px.Request{
		Method: pReq.GetMethod(),
		URL:    pReq.GetURL(),
		Proto:  pReq.GetProto(),
		Header: pReq.GetHeaders(),
	}
	body := pReq.GetBodyBytes()

	// Encode the body based on the accept encoding header
	encodedBody, encoding, err := utils.EncodeBody(body, acceptEncodingHeader)
	if err != nil {
		return nil, fmt.Errorf("error encoding body: %v", err)
	}

	// Set the new content encoding and length headers
	req.Header.Set("Content-Encoding", encoding)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

	req.Body = encodedBody
	return req, nil
}
