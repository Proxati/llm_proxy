package mitm

import (
	"fmt"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	px "github.com/proxati/mitmproxy/proxy"
)

// ToProxyRequest converts a RequestReaderAdapter into a MITM proxy request object with body encoding
// matching the provided acceptEncodingHeader. It encodes the request body based on the
// acceptEncodingHeader argument and sets the appropriate headers for content encoding and content length.
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
		return nil, fmt.Errorf("failed to encode request body with encoding '%s': %w", acceptEncodingHeader, err)

	}

	// Set the new content encoding and length headers
	if encoding != "" {
		req.Header.Set("Content-Encoding", encoding)
	} else {
		req.Header.Del("Content-Encoding")
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

	req.Body = encodedBody
	return req, nil
}
