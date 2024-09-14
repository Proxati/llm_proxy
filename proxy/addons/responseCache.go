package addons

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	px "github.com/proxati/mitmproxy/proxy"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons/cache"
	"github.com/proxati/llm_proxy/v2/proxy/addons/helpers"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/headers"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters/mitm"
	"github.com/proxati/llm_proxy/v2/schema/utils"
)

const (
	DefaultMemoryCacheSize = 1000 // number of records to cache per URL
)

var cacheOnlyMethods = map[string]struct{}{
	"GET":     {},
	"":        {},
	"HEAD":    {},
	"OPTIONS": {},
	"POST":    {},
}

var cacheOnlyResponseCodes = map[int]struct{}{
	http.StatusOK:        {},
	http.StatusAccepted:  {},
	http.StatusCreated:   {},
	http.StatusNoContent: {},
}

type ResponseCacheAddon struct {
	px.BaseAddon
	filterReqHeaders  *config.HeaderFilterGroup
	filterRespHeaders *config.HeaderFilterGroup
	formatter         formatters.MegaDumpFormatter
	cache             cache.DB
	wg                sync.WaitGroup
	closed            atomic.Bool
	logger            *slog.Logger
}

// requestOpen is the function typically used by the Request method
func (c *ResponseCacheAddon) requestOpen(logger *slog.Logger, f *px.Flow) {
	// Send a log message and set the Request header after this function returns.
	// The response header with cache status will be set in the Response method.
	// This is a bit of a hack because we don't have an easy way to store context.
	var cacheStatusHeaderValue string
	defer func() {
		if cacheStatusHeaderValue == "" {
			return
		}

		logger.Info("Cache", "status", cacheStatusHeaderValue)
		f.Request.Header.Set(headers.CacheStatusHeader, cacheStatusHeaderValue)
	}()

	// check if the request has a no-cache or no-store header, bypass the cache lookup if true
	cacheControlHeader := strings.ToLower(f.Request.Header.Get("Cache-Control"))
	for _, header := range []string{"no-cache", "no-store"} {
		if cacheControlHeader == header {
			logger.Debug(
				"skipping cache lookup because of the Cache-Control header value",
				"Cache-Control", header,
			)
			// hack to store the cache status in the request, the defer will set the header
			cacheStatusHeaderValue = headers.CacheStatusValueSkip
			return // return here to stop processing the rest of this function
		}
	}

	// Only cache these request methods (and empty string for GET)
	if _, ok := cacheOnlyMethods[f.Request.Method]; !ok {
		logger.Debug("skipping cache lookup for unsupported method", "method", f.Request.Method)
		cacheStatusHeaderValue = headers.CacheStatusValueSkip
		return
	}

	// decode the request body for cache lookup
	decodedBody, err := utils.DecodeBody(f.Request.Body, f.Request.Header.Get("Content-Encoding"))
	if err != nil {
		logger.Error("error decoding request body", "error", err)
		cacheStatusHeaderValue = headers.CacheStatusValueSkip
		return
	}

	// check the cache for responses matching this request
	cachedResponse, err := c.cache.Get(f.Request.URL.String(), decodedBody)
	if err != nil {
		logger.Error("error accessing cache, bypassing", "error", err)
		cacheStatusHeaderValue = headers.CacheStatusValueSkip
		return
	}

	// handle cache miss, return early otherwise NPEs below
	if cachedResponse == nil {
		cacheStatusHeaderValue = headers.CacheStatusValueMiss
		return
	}

	// handle cache hit
	cacheStatusHeaderValue = headers.CacheStatusValueHit

	// filter the headers before returning the cached response, and remove the Content-Encoding
	// header, because next we will be re-encoding the body according to the request's
	// Accept-Encoding header, and a new Content-Encoding header will be added.
	cachedResponse.Header = c.filterReqHeaders.FilterHeaders(
		cachedResponse.Header,
		// remove these headers from the cached response before returning, the CacheStatusHeader
		// will be re-added in the defer block
		"Content-Encoding", "Content-Length", headers.CacheStatusHeader,
	)

	// convert the cached response to a ProxyResponse, and encode the body according to the request's Accept-Encoding header
	encodedCachedResponse, err := mitm.ToProxyResponse(cachedResponse, f.Request.Header.Get("Accept-Encoding"))
	if err != nil {
		logger.Error("error converting cached response to ProxyResponse", "error", err)
		cacheStatusHeaderValue = headers.CacheStatusValueSkip
		return
	}

	// set the cache status header to indicate a hit
	encodedCachedResponse.Header.Set(headers.CacheStatusHeader, headers.CacheStatusValueHit)

	// other pending addons will be skipped after setting f.Response and returning from the caller method
	f.Response = encodedCachedResponse
}

func (c *ResponseCacheAddon) Request(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(c.logger, f).WithGroup("Request")
	c.wg.Add(1) // for blocking this addon during shutdown in .Close()
	defer c.wg.Done()

	if c.closed.Load() {
		helpers.RequestClosed(logger, f)
		return
	}

	c.requestOpen(logger, f)
}

// responseCommon is the function used by the Response method when the addon is both open or closed.
// It returns true if the addon should return early, and false otherwise.
func (c *ResponseCacheAddon) responseCommon(f *px.Flow) error {
	// Get the request header for headers.CacheStatusHeader and set the response header to whatever it's set to
	cacheStatus := f.Request.Header.Get(headers.CacheStatusHeader)

	if cacheStatus != "" {
		// Set the response header to the same value as the request header
		f.Response.Header.Set(headers.CacheStatusHeader, cacheStatus)
		if cacheStatus == headers.CacheStatusValueSkip {
			// not *really* an error, just a way to communicate the reason for skipping cache storage
			return fmt.Errorf("Cache header is set to: %s", cacheStatus)
		}
	}

	// Only cache good response codes
	_, shouldCache := cacheOnlyResponseCodes[f.Response.StatusCode]
	if !shouldCache {
		f.Response.Header.Set(headers.CacheStatusHeader, headers.CacheStatusValueSkip)
		return fmt.Errorf("response status code is not cacheable: %d", f.Response.StatusCode)
	}

	return nil
}

// responseStorage is the function used by the Response method (when the addon is open) to store
// the response in the cache. It will store the response in the cache, after filtering out the
// Content-Encoding and Content-Length headers. The lookup key is the request URL and the request
// body, and the cached value is the response object.
func (c *ResponseCacheAddon) responseStorage(f *px.Flow) error {
	// convert the request to an internal TrafficObject
	reqAdapter := mitm.NewProxyRequestAdapter(f.Request) // generic wrapper for the mitm request

	tObjReq, err := schema.NewProxyRequest(reqAdapter, c.filterReqHeaders)
	if err != nil {
		return fmt.Errorf("could not create TrafficObject from request: %w", err)
	}

	// convert the response to an internal TrafficObject
	respAdapter := mitm.NewProxyResponseAdapter(f.Response) // generic wrapper for the mitm response
	tObjResp, err := schema.NewProxyResponse(respAdapter, c.filterRespHeaders)
	if err != nil {
		return fmt.Errorf("could not create TrafficObject from response: %w", err)
	}

	// remove the Content-Encoding header to avoid storing this in the cache
	tObjResp.Header.Del("Content-Encoding")
	tObjResp.Header.Del("Content-Length")

	// store the response in the cache
	if err := c.cache.Put(tObjReq, tObjResp); err != nil {
		return fmt.Errorf("could not store response in cache: %w", err)
	}

	return nil
}

func (c *ResponseCacheAddon) Response(f *px.Flow) {
	c.wg.Add(1) // for blocking this addon during shutdown in .Close()
	logger := configLoggerFieldsWithFlow(c.logger, f).WithGroup("Response")

	earlyReturnErr := c.responseCommon(f)
	if earlyReturnErr != nil {
		logger.Debug("Skipping cache storage", "reason", earlyReturnErr.Error())
		c.wg.Done()
		return
	}

	if c.closed.Load() {
		logger.Warn("Skipping cache storage", "reason", "ResponseCacheAddon is being closed")
		c.wg.Done()
		return
	}

	go func() {
		logger.Debug("Response cache storage starting...")
		defer c.wg.Done() // .Done() must be inside the goroutine, so that .Close() waits for the storage to finish
		<-f.Done()        // block until the response from upstream is fully read into proxy memory, and other addons have run

		err := c.responseStorage(f)
		if err != nil {
			logger.Error("error storing response in cache", "error", err)
			return
		}
		logger.Debug("Response cache storage completed successfully")
	}()
}

func (d *ResponseCacheAddon) String() string {
	return fmt.Sprintf("ResponseCacheAddon (%s)", d.cache)
}

func (d *ResponseCacheAddon) Close() error {
	if !d.closed.Swap(true) {
		d.logger.Debug("Closing...")
		err := d.cache.Close()
		if err != nil {
			d.logger.Error("error closing cacheDB", "error", err)
		}
		d.logger.Debug("Waiting for any remaining cache storage operations to complete...")
		d.wg.Wait()
	}

	return nil
}

// cleanCacheDir ensures that the cache directory string is valid and returns the absolute path.
func cleanCacheDir(cacheDir string) (string, error) {
	if cacheDir == "" {
		cacheDir = "."
	}

	invalidChars := []string{"<", ">", ":", "\"", "\\", "|", "?", "*", "!", "+", "`", "'"}
	for _, char := range invalidChars {
		if strings.Contains(cacheDir, char) {
			return "", fmt.Errorf("filename contains invalid character: %s", char)
		}
	}

	cacheDir, err := filepath.Abs(cacheDir)
	if err != nil {
		return "", err
	}

	return cacheDir, nil
}

// NewCacheAddon creates a new ResponseCacheAddon.
//
// Parameters:
//   - logger: the DI'd logger
//   - storageEngineName: name of the storage engine to use
//   - cacheDir: output & cache storage directory
//   - filterReqHeaders: which headers to filter out from the request before logging
//   - filterRespHeaders: which headers to filter out from the response before logging
//
// Returns:
//
//	A new instance of ResponseCacheAddon (or error if one occurred)
func NewCacheAddon(
	logger *slog.Logger,
	storageEngineName string,
	cacheDir string,
	filterReqHeaders *config.HeaderFilterGroup,
	filterRespHeaders *config.HeaderFilterGroup,
) (*ResponseCacheAddon, error) {
	var cacheDB cache.DB
	var err error
	logger = logger.WithGroup("addons.ResponseCacheAddon")

	cacheDir, err = cleanCacheDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error cleaning cache path: %s", err)
	}

	switch storageEngineName {
	case "badger":
		// cacheDB, err = cache.NewBadgerMetaDB(cacheDir)
		panic("badger storage engine is disabled")
	case "bolt":
		// pass in the header filters for removing specific headers from the objects stored in cache
		cacheDB, err = cache.NewBoltMetaDB(logger, cacheDir)
		logger.Debug("Loaded BoltMetaDB database driver", "cacheDir", cacheDir)
	case "memory":
		cacheDB, err = cache.NewMemoryMetaDB(logger, DefaultMemoryCacheSize)
		logger.Debug("Loaded MemoryStorage database driver")
	default:
		return nil, fmt.Errorf("unknown storage engine: %s", storageEngineName)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating cache: %s", err)
	}

	if cacheDB == nil {
		return nil, fmt.Errorf("cacheDB is nil after initialization")
	}

	return &ResponseCacheAddon{
		formatter:         &formatters.JSON{},
		cache:             cacheDB,
		logger:            logger,
		closed:            atomic.Bool{},
		filterReqHeaders:  filterReqHeaders,
		filterRespHeaders: filterRespHeaders,
	}, nil
}
