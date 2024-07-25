package addons

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	px "github.com/proxati/mitmproxy/proxy"

	"github.com/proxati/llm_proxy/v2/proxy/addons/cache"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/schema"
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
	filterReqHeaders  []string
	filterRespHeaders []string
	formatter         formatters.MegaDumpFormatter
	cache             cache.DB
	closeOnce         sync.Once
	logger            *slog.Logger
}

func (c *ResponseCacheAddon) Request(f *px.Flow) {
	logger := c.logger.With("URL", f.Request.URL, "ID", f.Id.String())

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
	cacheLookup, err := c.cache.Get(f.Request.URL.String(), decodedBody)
	if err != nil {
		logger.Error("error accessing cache, bypassing", "error", err)
		return
	}

	// handle cache miss, return early otherwise NPEs below
	if cacheLookup == nil {
		f.Request.Header.Set(CacheStatusHeader, CacheStatusMiss)
		logger.Info("cache miss")
		return
	}

	// handle cache hit
	logger.Info("cache hit")

	cachedResp, err := cacheLookup.ToProxyResponse(f.Request.Header.Get("Accept-Encoding"))
	if err != nil {
		logger.Error("error converting cached response to ProxyResponse", "error", err)
		return
	}

	// set the cache status header to indicate a hit
	cacheLookup.Header.Set(CacheStatusHeader, CacheStatusHit)

	// other pending addons will be skipped after setting f.Response and returning from this method
	f.Response = cachedResp
}

func (c *ResponseCacheAddon) Response(f *px.Flow) {
	logger := c.logger.With("URL", f.Request.URL, "StatusCode", f.Response.StatusCode, "ID", f.Id.String())

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

	go func() {
		<-f.Done()
		// if the response is nil, don't even try to cache it
		if f.Response == nil {
			logger.Debug("skipping cache storage for nil response")
			return
		}

		// Only cache good response codes
		_, shouldCache := cacheOnlyResponseCodes[f.Response.StatusCode]
		if !shouldCache {
			f.Response.Header.Set(CacheStatusHeader, CacheStatusSkip)
			logger.Debug("skipping cache storage for non-200 response")
			return
		}

		// convert the request to an internal TrafficObject
		tObjReq, err := schema.NewProxyRequestFromMITMRequest(f.Request, c.filterReqHeaders)
		if err != nil {
			logger.Error("could not create TrafficObject from request", "error", err)
			return
		}
		// remove the Accept-Encoding header to avoid storing this in the cache
		tObjReq.Header.Del("Accept-Encoding")

		// convert the response to an internal TrafficObject
		tObjResp, err := schema.NewProxyResponseFromMITMResponse(f.Response, c.filterRespHeaders)
		if err != nil {
			logger.Error("could not create TrafficObject from response", "error", err)
			return
		}
		// remove the Content-Encoding header to avoid storing this in the cache
		tObjResp.Header.Del("Content-Encoding")

		if err := c.cache.Put(tObjReq, tObjResp); err != nil {
			logger.Error("could not store response in cache", "error", err)
		}

	}()
}

func (d *ResponseCacheAddon) String() string {
	return "ResponseCacheAddon"
}

func (d *ResponseCacheAddon) Close() (err error) {
	d.closeOnce.Do(func() {
		d.logger.Debug("Closing ResponseCacheAddon")
		err = d.cache.Close()
	})
	return
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

func NewCacheAddon(
	logger *slog.Logger,
	storageEngineName string, // name of the storage engine to use
	cacheDir string, // output & cache storage directory
	filterReqHeaders, filterRespHeaders []string, // which headers to filter out
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
		cacheDB, err = cache.NewBoltMetaDB(cacheDir, filterRespHeaders)
		logger.Debug("Loaded BoltMetaDB database driver", "cacheDir", cacheDir)
	default:
		return nil, fmt.Errorf("unknown storage engine: %s", storageEngineName)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating cache: %s", err)
	}

	return &ResponseCacheAddon{
		formatter: &formatters.JSON{},
		cache:     cacheDB,
		logger:    logger,
	}, nil
}
