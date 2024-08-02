package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
)

type ProxyResponse struct {
	Status       int         `json:"status,omitempty"`
	Header       http.Header `json:"header"`
	Body         string      `json:"body"`
	headerFilter *config.HeaderFilterGroup
}

func (pRes *ProxyResponse) GetStatusCode() int {
	return pRes.Status
}

func (pRes *ProxyResponse) GetHeaders() http.Header {
	return pRes.Header
}

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
		pRes.Header = header
		pRes.filterHeaders()
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

// NewFromMITMRequest creates a new ProxyRequest from a MITM proxy request object
func NewProxyResponse(req proxyAdapters.ResponseReaderAdapter, headerFilter *config.HeaderFilterGroup) (*ProxyResponse, error) {
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

// NewFromJSONBytes unmarshals a JSON object into a TrafficObject
func NewProxyResponseFromJSONBytes(data []byte, headerFilter *config.HeaderFilterGroup) (*ProxyResponse, error) {
	pRes := &ProxyResponse{}
	err := json.Unmarshal(data, pRes)
	if err != nil {
		return nil, err
	}

	if pRes.Header == nil {
		pRes.Header = make(http.Header)
	}

	pRes.headerFilter = headerFilter
	pRes.filterHeaders()

	return pRes, nil
}
