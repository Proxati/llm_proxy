package config

import (
	"net/http"
	"sync"
)

// persistentFiltersRequestLogs are headers that are always filtered from request logs
var persistentFiltersRequestLogs = []string{
	"Accept",
	"Accept-Encoding",
	"Connection",
}

// defaultFiltersRequestLogs are headers that are filtered from request logs by default, but can be
// overridden by the user
var defaultFiltersRequestLogs = []string{
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

// persistentFiltersResponseLogs are headers that are always filtered from response logs
var persistentFiltersResponseLogs = []string{
	"Connection",
	"Content-Length",
	"Content-Encoding",
}

// defaultFiltersResponseLogs are headers that are filtered from response logs by default, but can be
// overridden by the user
var defaultFiltersResponseLogs = []string{
	"Proxy-Authenticate",
	"Set-Cookie",
	"WWW-Authenticate",
}

const (
	// flagTitle_FilterRequestHeadersToLogs is the name of the filter group for headers to be filtered from requests to logs
	flagTitle_FilterRequestHeadersToLogs = "filter-request-headers-to-logs"

	// flagTitle_FilterResponseHeadersToLogs is the name of the filter group for headers to be filtered from responses to logs
	flagTitle_FilterResponseHeadersToLogs = "filter-response-headers-to-logs"

	// flagTitle_FilterRequestHeadersToUpstream is the name of the filter group for headers to be filtered from requests to upstream
	flagTitle_FilterRequestHeadersToUpstream = "filter-request-headers-to-upstream"

	// flagTitle_FilterResponseHeadersToClient is the name of the filter group for headers to be filtered from responses to the client
	flagTitle_FilterResponseHeadersToClient = "filter-response-headers-to-client"
)

type headerIndex map[string]any

// HeaderFilterGroup is an object used to filter headers for a specific purpose, such as when
// reading existing logs (remove content-type), writing new logs (remove auth), or when sending
// requests to upstream.
type HeaderFilterGroup struct {
	Headers []string
	index   headerIndex
	name    string
}

// NewHeaderFilterGroup creates a new HeaderFilterGroup
func NewHeaderFilterGroup(name string, headers []string) *HeaderFilterGroup {
	hfg := &HeaderFilterGroup{
		Headers: append([]string{}, headers...), // shallow copy the slice
		name:    name,
	}
	hfg.buildIndex()
	return hfg
}

func (hfg *HeaderFilterGroup) String() string {
	return hfg.name
}

func (hfg *HeaderFilterGroup) buildIndex() {
	index := make(headerIndex)
	for _, header := range hfg.Headers {
		index[header] = nil
	}
	hfg.index = index
}

// IsHeaderInGroup returns true if the header should be filtered by this group
func (hg *HeaderFilterGroup) IsHeaderInGroup(header string) bool {
	_, exists := hg.index[header]
	return exists
}

// FilterHeaders makes a shallow copy of the headers map and removes any headers that are in the
// filter group. additionalHeaders is a variadic parameter that allows for additional headers to be
// removed from the new map that will be returned by this method.
func (hg *HeaderFilterGroup) FilterHeaders(headers http.Header, additionalHeaders ...string) http.Header {
	filteredHeaders := make(http.Header)
	for header, values := range headers {
		if !hg.IsHeaderInGroup(header) {
			filteredHeaders[header] = values
		}
	}

	// Remove additional headers
	for _, header := range additionalHeaders {
		filteredHeaders.Del(header)
	}

	return filteredHeaders
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
		RequestToLogs: NewHeaderFilterGroup(
			flagTitle_FilterRequestHeadersToLogs, append(defaultFiltersRequestLogs, persistentFiltersRequestLogs...)),
		ResponseToLogs: NewHeaderFilterGroup(
			flagTitle_FilterResponseHeadersToLogs, append(defaultFiltersResponseLogs, persistentFiltersResponseLogs...)),
		RequestToUpstream: NewHeaderFilterGroup(
			flagTitle_FilterRequestHeadersToUpstream, []string{}),
		ResponseToClient: NewHeaderFilterGroup(
			flagTitle_FilterResponseHeadersToClient, []string{}),
	}
	return hfc
}

// BuildIndexes rebuilds the internal indexes for the header filters
func (hg *HeaderFiltersContainer) BuildIndexes() {
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		defer wg.Done()
		hg.RequestToLogs.buildIndex()
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
