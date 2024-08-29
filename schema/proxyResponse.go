package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
)

// ProxyResponse is a struct that represents a response from a proxied request
type ProxyResponse struct {
	Status       int         `json:"status,omitempty"`
	Header       http.Header `json:"header"`
	Body         string      `json:"body"`
	headerFilter *config.HeaderFilterGroup
}

// GetStatusCode returns the status code in the ProxyResponse object
func (pRes *ProxyResponse) GetStatusCode() int {
	return pRes.Status
}

// GetHeaders returns the headers in the ProxyResponse object
func (pRes *ProxyResponse) GetHeaders() http.Header {
	return pRes.Header
}

// GetBodyBytes returns the body as a byte slice
func (pRes *ProxyResponse) GetBodyBytes() []byte {
	return []byte(pRes.Body)
}

// filterHeaders filters the headers in the ProxyResponse object using the headerFilter object
func (pRes *ProxyResponse) filterHeaders() {
	if pRes.headerFilter == nil {
		return
	}
	pRes.Header = pRes.headerFilter.FilterHeaders(pRes.Header)
}

// loadBody loads the request body into the ProxyRequest object
func (pRes *ProxyResponse) loadBody(body []byte, contentEncoding string) error {
	var bodyIsPrintable bool

	decodedBody, err := utils.DecodeBody(body, contentEncoding)
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
			// don't load the header if it's in the filter group
			if pRes.headerFilter != nil && pRes.headerFilter.IsHeaderInGroup(k) {
				continue
			}

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
		// store headers
		pRes.Header = header
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

// NewProxyResponse creates a new ProxyRequest from a MITM proxy request object
func NewProxyResponse(req proxyadapters.ResponseReaderAdapter, headerFilter *config.HeaderFilterGroup) (*ProxyResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("response is nil, unable to create ProxyResponse")
	}

	pRes := &ProxyResponse{
		Status:       req.GetStatusCode(),
		Header:       req.GetHeaders(),
		headerFilter: headerFilter,
	}
	if err := pRes.loadBody(req.GetBodyBytes(), pRes.Header.Get("Content-Encoding")); err != nil {
		getLogger().Warn("could not load ProxyResponse body", "error", err)
		pRes.Body = ""
	}

	pRes.filterHeaders()
	return pRes, nil
}

// NewProxyResponseFromJSONBytes unmarshals a JSON object into a TrafficObject.
// Headers must be filtered later.
func NewProxyResponseFromJSONBytes(data []byte) (*ProxyResponse, error) {
	pRes := &ProxyResponse{}
	err := json.Unmarshal(data, pRes)
	if err != nil {
		return nil, err
	}

	if pRes.Header == nil {
		pRes.Header = make(http.Header)
	}

	return pRes, nil
}
