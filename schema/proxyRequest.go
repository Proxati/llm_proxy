package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	px "github.com/proxati/mitmproxy/proxy"

	"github.com/proxati/llm_proxy/v2/schema/utils"
)

type ProxyRequest struct {
	Method            string         `json:"method,omitempty"`
	URL               *url.URL       `json:"url,omitempty"`
	Proto             string         `json:"proto,omitempty"`
	Header            http.Header    `json:"header"`
	Body              string         `json:"body"`
	headerFilterIndex map[string]any `json:"-"`
}

// loadHeaderFilterIndex loads the headers to filter into a map, used by loadHeaders
func (pReq *ProxyRequest) loadHeaderFilterIndex(headersToFilter []string) {
	pReq.headerFilterIndex = make(map[string]any)
	for _, header := range headersToFilter {
		pReq.headerFilterIndex[header] = nil
	}
}

// loadHeaders resets and loads the new headers into the ProxyRequest object
func (pReq *ProxyRequest) loadHeaders(headers map[string][]string) {
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
func (pReq *ProxyRequest) loadBody(body []byte) error {
	var bodyIsPrintable bool

	pReq.Body, bodyIsPrintable = utils.CanPrintFast(body)
	if !bodyIsPrintable {
		return errors.New("request body is not printable")
	}

	return nil
}

// HeaderString returns the headers as a flat string
func (pReq *ProxyRequest) HeaderString() string {
	return utils.HeaderString(pReq.Header)
}

// UnmarshalJSON performs a non-threadsafe load of json data into THIS ProxyRequest
func (pReq *ProxyRequest) UnmarshalJSON(data []byte) error {
	r := make(map[string]any)
	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	// handle method
	method, ok := r["method"]
	if ok {
		pReq.Method = method.(string)
	}

	// handle URL
	rawURL, ok := r["url"]
	if ok {
		strURL, ok := rawURL.(string)
		if !ok {
			return errors.New("url parse error")
		}
		u, err := url.Parse(strURL)
		if err != nil {
			return err
		}
		pReq.URL = u
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
		// load and filter headers
		pReq.loadHeaders(header)
	}

	// handle body
	body, ok := r["body"]
	if ok {
		pReq.Body, ok = body.(string)
		if !ok {
			return errors.New("body parse error")
		}
	}

	// handle proto
	proto, ok := r["proto"]
	if ok {
		pReq.Proto, ok = proto.(string)
		if !ok {
			return errors.New("proto parse error")
		}
	}
	return nil
}

// MarshalJSON dumps this ProxyRequest into a byte array containing JSON
func (pReq *ProxyRequest) MarshalJSON() ([]byte, error) {
	var urlString string
	if pReq.URL != nil {
		urlString = pReq.URL.String()
	}

	type Alias ProxyRequest
	return json.Marshal(&struct {
		URL string `json:"url,omitempty"`
		*Alias
	}{
		URL:   urlString,
		Alias: (*Alias)(pReq),
	})
}

// NewFromMITMRequest creates a new ProxyRequest from a MITM proxy request object
func NewProxyRequestFromMITMRequest(req *px.Request, headersToFilter []string) (*ProxyRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil, unable to create ProxyRequest")
	}

	pReq := &ProxyRequest{
		Method: req.Method,
		URL:    req.URL,
		Proto:  req.Proto,
	}

	pReq.loadHeaderFilterIndex(headersToFilter)
	pReq.loadHeaders(req.Header)
	if err := pReq.loadBody(req.Body); err != nil {
		if req.URL != nil {
			getLogger().Warn("unable to load request body", "URL", req.URL.String())
		} else {
			getLogger().Warn("unable to load request body")
		}
	}

	return pReq, nil
}
