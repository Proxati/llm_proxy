package config

import (
	"log/slog"
	"net/http"
	"sync"
)

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

var defaultFiltersResponseLogs = []string{
	"Proxy-Authenticate",
	"Set-Cookie",
	"WWW-Authenticate",
}

const (
	// FlagTitle_FilterRequestHeadersToLogs is the name of the filter group for headers to be filtered from requests to logs
	FlagTitle_FilterRequestHeadersToLogs = "filter-request-headers-to-logs"

	// FlagTitle_FilterResponseHeadersToLogs is the name of the filter group for headers to be filtered from responses to logs
	FlagTitle_FilterResponseHeadersToLogs = "filter-response-headers-to-logs"

	// FlagTitle_FilterRequestHeadersToUpstream is the name of the filter group for headers to be filtered from requests to upstream
	FlagTitle_FilterRequestHeadersToUpstream = "filter-request-headers-to-upstream"

	// FlagTitle_FilterResponseHeadersToClient is the name of the filter group for headers to be filtered from responses to the client
	FlagTitle_FilterResponseHeadersToClient = "filter-response-headers-to-client"
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
			FlagTitle_FilterRequestHeadersToLogs, defaultFiltersRequestLogs),
		ResponseToLogs: NewHeaderFilterGroup(
			FlagTitle_FilterResponseHeadersToLogs, defaultFiltersResponseLogs),
		RequestToUpstream: NewHeaderFilterGroup(
			FlagTitle_FilterRequestHeadersToUpstream, []string{}),
		ResponseToClient: NewHeaderFilterGroup(
			FlagTitle_FilterResponseHeadersToClient, []string{}),
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
