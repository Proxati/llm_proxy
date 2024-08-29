package config

import (
	"net/http"
	"sync"
)

// persistentFiltersRequests are headers that are always filtered from all request logs
var persistentFiltersRequests = []string{
	"Accept",
	"Accept-Encoding",
	"Connection",
}

// defaultFiltersRequests are headers that are filtered from request logs by default, but can be
// overridden by the user
var defaultFiltersRequests = []string{
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

// persistentFiltersResponses are headers that are always filtered from response logs
var persistentFiltersResponses = []string{
	"Connection",
	"Content-Length",
	"Content-Encoding",
}

// defaultFiltersResponses are headers that are filtered from response logs by default, but can be
// overridden by the user
var defaultFiltersResponses = []string{
	"Proxy-Authenticate",
	"Set-Cookie",
	"WWW-Authenticate",
}

const (
	flagTitleFilterResponseHeadersToCache   = "filter-response-headers-to-cache"
	flagTitleFilterRequestHeadersToLogs     = "filter-request-headers-to-logs"
	flagTitleFilterResponseHeadersToLogs    = "filter-response-headers-to-logs"
	flagTitleFilterRequestHeadersToUpstream = "filter-request-headers-to-upstream"
	flagTitleFilterResponseHeadersToClient  = "filter-response-headers-to-client"
)

type headerIndex map[string]any

// HeaderFilterGroup is an object used to filter headers for a specific purpose, such as when
// reading existing logs (remove content-type), writing new logs (remove auth), or when sending
// requests to upstream.
type HeaderFilterGroup struct {
	Headers           []string    // user-editable list of headers to filter
	additionalHeaders []string    // headers that are always filtered, not user editable
	index             headerIndex // map of headers to filter
	name              string      // human-readable name of this filter group
}

// NewHeaderFilterGroup creates a new HeaderFilterGroup
func NewHeaderFilterGroup(name string, userHeaders []string, extraHeaders ...[]string) *HeaderFilterGroup {
	var mergedHeaders []string
	for _, headers := range extraHeaders {
		mergedHeaders = append(mergedHeaders, headers...)
	}

	hfg := &HeaderFilterGroup{
		Headers:           append([]string{}, userHeaders...), // shallow copy the slice
		additionalHeaders: mergedHeaders,
		name:              name,
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
	for _, header := range hfg.additionalHeaders {
		index[header] = nil
	}
	hfg.index = index
}

// IsHeaderInGroup returns true if the header should be filtered by this group
func (hfg *HeaderFilterGroup) IsHeaderInGroup(header string) bool {
	_, exists := hfg.index[header]
	return exists
}

// FilterHeaders makes a shallow copy of the headers map and removes any headers that are in the
// filter group. additionalHeaders is a variadic parameter that allows for additional headers to be
// removed from the new map that will be returned by this method.
func (hfg *HeaderFilterGroup) FilterHeaders(headers http.Header, additionalHeaders ...string) http.Header {
	filteredHeaders := make(http.Header)
	for header, values := range headers {
		if !hfg.IsHeaderInGroup(header) {
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
	// Headers set by the upstream, to be filtered from requests stored in the cache
	ResponseToCache *HeaderFilterGroup

	// Headers set by the client, to be omitted from logs
	RequestToLogs *HeaderFilterGroup // filter-request-headers-to-logs

	// Headers set by upstream, to be omitted from logs
	ResponseToLogs *HeaderFilterGroup // filter-response-headers-to-logs

	// Headers set by the client, to be filtered from requests sent upstream
	RequestToUpstream *HeaderFilterGroup // filter-request-headers-to-upstream

	// Headers set by upstream, to be filtered from the responses sent to the client
	ResponseToClient *HeaderFilterGroup // filter-response-headers-to-client
}

// NewHeaderFiltersContainer creates a new HeaderFiltersContainer with default values
func NewHeaderFiltersContainer() *HeaderFiltersContainer {
	hfc := &HeaderFiltersContainer{
		ResponseToCache: NewHeaderFilterGroup(
			flagTitleFilterResponseHeadersToCache,
			defaultFiltersResponses,
			persistentFiltersResponses,
		),
		RequestToLogs: NewHeaderFilterGroup(
			flagTitleFilterRequestHeadersToLogs,
			defaultFiltersRequests,
			persistentFiltersRequests,
		),
		ResponseToLogs: NewHeaderFilterGroup(
			flagTitleFilterResponseHeadersToLogs,
			defaultFiltersResponses,
			persistentFiltersResponses,
		),
		RequestToUpstream: NewHeaderFilterGroup(
			flagTitleFilterRequestHeadersToUpstream, []string{}, []string{}),
		ResponseToClient: NewHeaderFilterGroup(
			flagTitleFilterResponseHeadersToClient, []string{}, []string{}),
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
