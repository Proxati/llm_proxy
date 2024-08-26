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
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters/mitm"
	"github.com/proxati/llm_proxy/v2/schema/utils"
)

const (
	CacheStatusHeader      = "X-Llm_proxy-Cache"
	CacheStatusHit         = "HIT"
	CacheStatusMiss        = "MISS"
	CacheStatusSkip        = "SKIP"
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

func (c *ResponseCacheAddon) Request(f *px.Flow) {
	logger := c.logger.With(
		"client.ID", f.Id.String(),
		"proxy.ID", f.ConnContext.ID(),
		"URL", f.Request.URL,
		"Method", f.Request.Method,
	)

	if c.closed.Load() {
		logger.Warn("ResponseCacheAddon is being closed, skipping request")
		f.Response = &px.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       []byte("LLM_Proxy is not available"),
			Header: http.Header{
				"Content-Type":    {"text/plain"},
				CacheStatusHeader: {CacheStatusSkip},
			},
		}
		return
	}

	// wg incremented after the closed check, otherwise close may wait forever as new requests pile up
	c.wg.Add(1)
	defer c.wg.Done()

	// check if the request has a no-cache or no-store header, bypass the cache lookup if true
	cacheControlHeader := strings.ToLower(f.Request.Header.Get("Cache-Control"))
	for _, header := range []string{"no-cache", "no-store"} {
		if cacheControlHeader == header {
			logger.Debug(
				"skipping cache lookup because of the Cache-Control header value",
				"Cache-Control", header,
			)
			f.Request.Header.Set(CacheStatusHeader, CacheStatusSkip) // hack to store the cache status in the request
			return
		}
	}

	// Only cache these request methods (and empty string for GET)
	if _, ok := cacheOnlyMethods[f.Request.Method]; !ok {
		logger.Debug("skipping cache lookup for unsupported method", "method", f.Request.Method)
		f.Request.Header.Set(CacheStatusHeader, CacheStatusSkip) // hack to store the cache status state in the request
		return
	}

	// decode the request body for cache lookup
	decodedBody, err := utils.DecodeBody(f.Request.Body, f.Request.Header.Get("Content-Encoding"))
	if err != nil {
		logger.Error("error decoding request body", "error", err)
		return
	}

	// check the cache for responses matching this request
	cachedResponse, err := c.cache.Get(f.Request.URL.String(), decodedBody)
	if err != nil {
		logger.Error("error accessing cache, bypassing", "error", err)
		return
	}

	// handle cache miss, return early otherwise NPEs below
	if cachedResponse == nil {
		f.Request.Header.Set(CacheStatusHeader, CacheStatusMiss)
		logger.Info("cache miss")
		return
	}

	// handle cache hit
	logger.Info("cache hit")

	// filter the headers before returning the cached response, and remove the Content-Encoding
	// header, because next we will be re-encoding the body according to the request's
	// Accept-Encoding header, and a new Content-Encoding header will be added.
	cachedResponse.Header = c.filterReqHeaders.FilterHeaders(
		cachedResponse.Header, "Content-Encoding", "Content-Length")

	// convert the cached response to a ProxyResponse, and encode the body according to the request's Accept-Encoding header
	encodedCachedResponse, err := mitm.ToProxyResponse(cachedResponse, f.Request.Header.Get("Accept-Encoding"))
	if err != nil {
		logger.Error("error converting cached response to ProxyResponse", "error", err)
		return
	}

	// set the cache status header to indicate a hit
	encodedCachedResponse.Header.Set(CacheStatusHeader, CacheStatusHit)

	// other pending addons will be skipped after setting f.Response and returning from this method
	f.Response = encodedCachedResponse
}

func (c *ResponseCacheAddon) Response(f *px.Flow) {
	c.wg.Add(1) // for blocking this addon during shutdown in .Close()
	// Get the request header for CacheStatusHeader and set the response header to whatever it's set to
	cacheStatus := f.Request.Header.Get(CacheStatusHeader)

	logger := c.logger.With(
		"client.ID", f.Id.String(),
		"proxy.ID", f.ConnContext.ID(),
		"StatusCode", f.Response.StatusCode,
		"URL", f.Request.URL,
		"Method", f.Request.Method,
	)

	if cacheStatus != "" {
		logger = logger.With("CacheStatus", cacheStatus)
		// Set the response header to the same value as the request header
		f.Response.Header.Set(CacheStatusHeader, cacheStatus)
		if cacheStatus == CacheStatusSkip {
			logger.Debug("skipping cache storage")
			c.wg.Done()
			return
		}
	}

	// Only cache good response codes
	_, shouldCache := cacheOnlyResponseCodes[f.Response.StatusCode]
	if !shouldCache {
		f.Response.Header.Set(CacheStatusHeader, CacheStatusSkip)
		logger.Debug("skipping cache storage for non-200 response")
		c.wg.Done()
		return
	}

	if c.closed.Load() {
		logger.Warn("ResponseCacheAddon is being closed, not storing response in cache")
		c.wg.Done()
		return
	}

	go func() {
		logger.Debug("Response cache storage starting...")
		defer c.wg.Done() // .Done() must be inside the goroutine, so that .Close() waits for the storage to finish
		<-f.Done()        // block until the response from upstream is fully read into proxy memory, and other addons have run

		// convert the request to an internal TrafficObject
		reqAdapter := mitm.NewProxyRequestAdapter(f.Request) // generic wrapper for the mitm request

		tObjReq, err := schema.NewProxyRequest(reqAdapter, c.filterReqHeaders)
		if err != nil {
			logger.Error("could not create TrafficObject from request", "error", err)
			return
		}

		// remove the Accept-Encoding header to avoid storing this in the cache
		originalAcceptEncoding := tObjReq.Header.Get("Accept-Encoding")

		// convert the response to an internal TrafficObject
		respAdapter := mitm.NewProxyResponseAdapter(f.Response) // generic wrapper for the mitm response
		tObjResp, err := schema.NewProxyResponse(respAdapter, c.filterRespHeaders)
		if err != nil {
			logger.Error("could not create TrafficObject from response", "error", err)
			return
		}

		// remove the Content-Encoding header to avoid storing this in the cache
		tObjResp.Header.Del("Content-Encoding")
		tObjResp.Header.Del("Content-Length")
		logger.Debug(
			"removed header from request before storing in cache",
			"Accept-Encoding", originalAcceptEncoding,
		)

		// store the response in the cache
		if err := c.cache.Put(tObjReq, tObjResp); err != nil {
			logger.Error("could not store response in cache", "error", err)
		}

		logger.Debug("Response cache storage complete")
	}()
}

func (d *ResponseCacheAddon) String() string {
	return "ResponseCacheAddon"
}

func (d *ResponseCacheAddon) Close() error {
	if !d.closed.Swap(true) {
		d.logger.Debug("Closing...")
		err := d.cache.Close()
		if err != nil {
			d.logger.Error("error closing cacheDB", "error", err)
		}
		d.wg.Wait()
	}

	return nil
}

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
		cacheDB, err = cache.NewBoltMetaDB(cacheDir)
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
