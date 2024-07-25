package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"

	"github.com/proxati/llm_proxy/v2/schema/utils"
)

type ProxyResponse struct {
	Status            int            `json:"status,omitempty"`
	Header            http.Header    `json:"header"`
	Body              string         `json:"body"`
	headerFilterIndex map[string]any `json:"-"`
}

// loadHeaderFilterIndex loads the headers to filter into a map, used by loadHeaders
func (pReq *ProxyResponse) loadHeaderFilterIndex(headersToFilter []string) {
	pReq.headerFilterIndex = make(map[string]any)
	for _, header := range headersToFilter {
		pReq.headerFilterIndex[header] = nil
	}
}

// loadHeaders resets and loads the new headers into the ProxyRequest object
func (pReq *ProxyResponse) loadHeaders(headers map[string][]string) {
	pReq.Header = make(http.Header)
	if pReq.headerFilterIndex == nil {
		pReq.Header = headers
		return // no headers to filter
	}

	for key, values := range headers {
		if _, found := pReq.headerFilterIndex[key]; !found {
			for _, value := range values {
				pReq.Header.Add(key, value)
			}
		}
	}
}

// loadBody loads the request body into the ProxyRequest object
func (pRes *ProxyResponse) loadBody(body []byte, content_encoding string) error {
	var bodyIsPrintable bool

	decodedBody, err := utils.DecodeBody(body, content_encoding)
	if err != nil {
		return fmt.Errorf("error decoding body: %w", err)
	}

	pRes.Body, bodyIsPrintable = utils.CanPrintFast(decodedBody)
	if !bodyIsPrintable {
		return errors.New("response body is not printable")
	}

	return nil
}

// HeaderString returns the headers as a flat string
func (pRes *ProxyResponse) HeaderString() string {
	return utils.HeaderString(pRes.Header)
}

// UnmarshalJSON performs a non-threadsafe load of json data into THIS ProxyResponse
func (pRes *ProxyResponse) UnmarshalJSON(data []byte) error {
	r := make(map[string]any)
	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	// handle status code
	if statusCode, ok := r["status"]; ok {
		statusFloat, ok := statusCode.(float64)
		if !ok {
			return errors.New("status parse error")
		}
		pRes.Status = int(statusFloat)
	}

	// handle headers
	rawheader, ok := r["header"].(map[string]any)
	if ok {
		header := make(map[string][]string)
		for k, v := range rawheader {
			vals, ok := v.([]any)
			if !ok {
				return errors.New("header parse error")
			}

			svals := make([]string, 0)
			for _, val := range vals {
				sval, ok := val.(string)
				if !ok {
					return errors.New("header parse error")
				}
				svals = append(svals, sval)
			}
			header[k] = svals
		}
		pRes.loadHeaders(header)
	}

	// handle body
	body, ok := r["body"]
	if ok {
		pRes.Body, ok = body.(string)
		if !ok {
			return errors.New("body parse error")
		}
	}

	return nil
}

// ToProxyResponse converts a ProxyResponse into a MITM proxy response object (with content encoding matching the new req)
// Because all responses are stored as uncompressed strings, the cached response might need to be encoded before being sent
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

// NewFromMITMRequest creates a new ProxyRequest from a MITM proxy request object
func NewProxyResponseFromMITMResponse(req *px.Response, headersToFilter []string) (*ProxyResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("response is nil, unable to create ProxyResponse")
	}

	pRes := &ProxyResponse{
		Status: req.StatusCode,
	}

	pRes.loadHeaderFilterIndex(headersToFilter)
	pRes.loadHeaders(req.Header)

	if err := pRes.loadBody(req.Body, req.Header.Get("Content-Encoding")); err != nil {
		getLogger().Warn("could not load ProxyResponse body", "error", err)
	}

	return pRes, nil
}

// NewFromJSONBytes unmarshals a JSON object into a TrafficObject
func NewProxyResponseFromJSONBytes(data []byte, headersToFilter []string) (*ProxyResponse, error) {
	pRes := &ProxyResponse{}
	pRes.loadHeaderFilterIndex(headersToFilter)

	err := json.Unmarshal(data, pRes)
	if err != nil {
		return nil, err
	}
	pRes.loadHeaders(pRes.Header)

	return pRes, nil
}
