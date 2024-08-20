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
	CacheStatusHeader = "X-Llm_proxy-Cache"
	CacheStatusHit    = "HIT"
	CacheStatusMiss   = "MISS"
	CacheStatusSkip   = "SKIP"
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
	logger := c.logger.With("URL", f.Request.URL, "ID", f.Id.String())

	if c.closed.Load() {
		logger.Warn("ResponseCacheAddon is being closed, skipping request")
		f.Response = &px.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       []byte("LLM_Proxy is not available"),
		}
		f.Response.Header.Set(CacheStatusHeader, CacheStatusSkip)
		return
	}

	c.wg.Add(1) // for blocking this addon during shutdown in .Close()
	defer c.wg.Done()

	if f.Request.URL == nil || f.Request.URL.String() == "" {
		logger.Error("request URL is nil or empty")
		f.Response = &px.Response{
			StatusCode: http.StatusBadRequest,
			Body:       []byte("Request URL is empty"),
		}
		return
	}

	// Only cache these request methods (and empty string for GET)
	if _, ok := cacheOnlyMethods[f.Request.Method]; !ok {
		logger.Debug("skipping cache lookup for unsupported method", "method", f.Request.Method)
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
	cachedResponse.Header = c.filterReqHeaders.FilterHeaders(cachedResponse.Header, "Content-Encoding", "Content-Length")

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
	logger := c.logger.With(
		"URL", f.Request.URL,
		"StatusCode", f.Response.StatusCode,
		"ID", f.Id.String(),
	)

	// if the response is nil, don't even try to cache it
	if f.Response == nil {
		logger.Debug("skipping cache storage for nil response")
		return
	}

	// add a header to the response to indicate it was a cache miss
	if f.Request != nil && f.Request.Header.Get(CacheStatusHeader) == CacheStatusMiss {
		// abusing the request header as a context storage for the cache miss
		if f.Response != nil {
			f.Response.Header.Set(CacheStatusHeader, CacheStatusMiss)
		}
	}

	if c.closed.Load() {
		logger.Warn("ResponseCacheAddon is being closed, not storing response in cache")
		return
	}

	c.wg.Add(1) // for blocking this addon during shutdown in .Close()
	go func() {
		logger.Debug("Request starting...")
		defer c.wg.Done()
		<-f.Done()

		// Only cache good response codes
		_, shouldCache := cacheOnlyResponseCodes[f.Response.StatusCode]
		if !shouldCache {
			f.Response.Header.Set(CacheStatusHeader, CacheStatusSkip)
			logger.Debug("skipping cache storage for non-200 response")
			return
		}

		// convert the request to an internal TrafficObject
		reqAdapter := mitm.NewProxyRequestAdapter(f.Request) // generic wrapper for the mitm request

		tObjReq, err := schema.NewProxyRequest(reqAdapter, c.filterReqHeaders)
		if err != nil {
			logger.Error("could not create TrafficObject from request", "error", err)
			return
		}

		// remove the Accept-Encoding header to avoid storing this in the cache
		originalAcceptEncoding := tObjReq.Header.Get("Accept-Encoding")
		logger.Debug("removing Accept-Encoding header from request", "Accept-Encoding", originalAcceptEncoding)

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

		// store the response in the cache
		if err := c.cache.Put(tObjReq, tObjResp); err != nil {
			logger.Error("could not store response in cache", "error", err)
		}

	}()
}

func (d *ResponseCacheAddon) String() string {
	return "ResponseCacheAddon"
}

func (d *ResponseCacheAddon) Close() error {
	if !d.closed.Swap(true) {
		d.logger.Debug("Closing ResponseCacheAddon")
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
		cacheDB, err = cache.NewMemoryMetaDB(logger, 1000)
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
