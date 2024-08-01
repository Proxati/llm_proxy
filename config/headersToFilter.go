package config

import (
	"log/slog"
	"net/http"
	"sync"
)

var defaultRequestFilterHeaders = []string{
	"Accept-Encoding",
	"Authorization",
	"Authorization-Info",
	"Cookie",
	"Openai-Organization",
	"Proxy-Authorization",
	"X-Access-Token",
	"X-Api-Key",
	"X-Auth-Password",
	"X-Auth-Token",
	"X-Auth-User",
	"X-CSRF-Token",
	"X-Forwarded-User",
	"X-Password",
	"X-Refresh-Token",
	"X-User-Email",
	"X-User-Id",
	"X-UserName",
	"X-User-Secret",
}

var defaultResponseFilterHeaders = []string{
	"Proxy-Authenticate",
	"Set-Cookie",
	"WWW-Authenticate",
}

type headerIndex map[string]any

type HeaderFilterGroup struct {
	Headers []string
	index   headerIndex
}

func NewHeaderFilterGroup(headers []string) *HeaderFilterGroup {
	hfg := &HeaderFilterGroup{
		Headers: headers,
	}
	hfg.buildIndex()
	return hfg
}

func (hfg *HeaderFilterGroup) buildIndex() {
	index := make(headerIndex)
	for _, header := range hfg.Headers {
		index[header] = nil
	}
	hfg.index = index
}

func (hg *HeaderFilterGroup) IsHeaderInGroup(header string) bool {
	_, exists := hg.index[header]
	return exists
}

func (hg *HeaderFilterGroup) FilterHeaders(headers http.Header) http.Header {
	for header := range headers {
		if hg.IsHeaderInGroup(header) {
			headers.Del(header)
		}
	}
	return headers
}

// HeaderFiltersContainer holds the configuration for filtering headers
type HeaderFiltersContainer struct {
	// Headers to be ignored by the proxy when received by the client
	RequestToLogs *HeaderFilterGroup // filter-request-headers-to-logs

	// Headers to be ignored by the proxy from responses from upstream
	ResponseToLogs *HeaderFilterGroup // filter-response-headers-to-logs

	// Headers to be filtered from requests coming from upstream
	RequestToUpstream *HeaderFilterGroup // filter-request-headers-to-upstream

	// Headers to be filtered from responses coming from upstream
	ResponseToClient *HeaderFilterGroup // filter-response-headers-to-client
}

// NewHeaderFiltersContainer creates a new HeaderFiltersContainer with default values
func NewHeaderFiltersContainer() *HeaderFiltersContainer {
	hfc := &HeaderFiltersContainer{
		RequestToLogs: &HeaderFilterGroup{
			Headers: append([]string{}, defaultRequestFilterHeaders...),
		},
		ResponseToLogs: &HeaderFilterGroup{
			Headers: append([]string{}, defaultResponseFilterHeaders...),
		},
		RequestToUpstream: &HeaderFilterGroup{},
		ResponseToClient:  &HeaderFilterGroup{},
	}
	hfc.BuildIndexes()

	return hfc
}

// BuildIndexes rebuilds the internal indexes for the header filters
func (hg *HeaderFiltersContainer) BuildIndexes() {
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		defer wg.Done()
		hg.RequestToLogs.buildIndex()
		if !hg.RequestToLogs.IsHeaderInGroup("Accept-Encoding") {
			slog.Default().Warn(
				"Accept-Encoding header is not in the filter-request-headers-to-logs filter group. This may cause issues with the cache.")
		}
	}()
	go func() {
		defer wg.Done()
		hg.ResponseToLogs.buildIndex()
	}()
	go func() {
		defer wg.Done()
		hg.RequestToUpstream.buildIndex()
	}()
	go func() {
		defer wg.Done()
		hg.ResponseToClient.buildIndex()
	}()
	wg.Wait()
}
