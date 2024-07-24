package addons

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	px "github.com/proxati/mitmproxy/proxy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanCachePath(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("")
		require.Nil(t, err)
		assert.Equal(t, currentDir, cacheDir)
	})

	t.Run(". cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir(".")
		require.Nil(t, err)
		assert.Equal(t, currentDir, cacheDir)
	})

	t.Run("non-empty cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("/tmp")
		require.Nil(t, err)
		assert.Equal(t, "/tmp", cacheDir)
	})

	t.Run("invalid cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("\\\\invalid\\path")
		require.NotNil(t, err)
		assert.Equal(t, "", cacheDir)
	})

	t.Run("relative cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("../../../../../../../../../../../tmp")
		assert.Nil(t, err)
		assert.Equal(t, "/tmp", cacheDir)
	})
}

func TestNewCacheAddonErr(t *testing.T) {
	testLogger := slog.Default()
	filterReqHeaders := []string{"header1", "header2"}
	filterRespHeaders := []string{"header1", "header2"}

	t.Run("empty storage engine", func(t *testing.T) {
		storageEngineName := ""
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.Error(t, err, "Expected error for empty storage engine")
		assert.Nil(t, cache)
	})

	t.Run("unknown storage engine", func(t *testing.T) {
		storageEngineName := "unknown"
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.Error(t, err, "Expected error for unknown storage engine")
		assert.Nil(t, cache)
	})

	t.Run("bolt storage engine with invalid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := "\\\\invalid\\path"
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.Error(t, err, "Expected error for invalid cacheDir")
		assert.Nil(t, cache)
	})

	t.Run("bolt storage engine with valid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.NoError(t, err, "Expected no error for valid cacheDir")
		assert.NotNil(t, cache)
		assert.Equal(t, "ResponseCacheAddon", cache.String())
	})
}

func TestRequest(t *testing.T) {
	testLogger := slog.Default()
	tmpDir := t.TempDir()
	filterReqHeaders := []string{"Header1"}
	filterRespHeaders := []string{"Header2"}
	respCacheAddon, err := NewCacheAddon(
		testLogger,
		"bolt", tmpDir,
		filterReqHeaders, filterRespHeaders,
	)
	require.Nil(t, err, "No error creating cache addon")

	t.Run("cache miss", func(t *testing.T) {
		// first request made with an empty request body
		flow := &px.Flow{
			Request: &px.Request{
				Method: "POST",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":    []string{"example.com"},
					"header1": []string{"value1"},
				},
				Body: []byte("req"),
			},
		}
		require.Empty(t, flow.Request.Header.Get(CacheStatusHeader))

		// simulate the request hitting the addon
		respCacheAddon.Request(flow)
		require.Nil(t, flow.Response, "nil response means cache miss")
		require.NotEmpty(t, flow.Request.Header.Get(CacheStatusHeader), "expected X-Cache header to exist")
		assert.Equal(t, "MISS", flow.Request.Header.Get(CacheStatusHeader), "expected X-Cache value to be MISS")
	})

	t.Run("cache hit", func(t *testing.T) {
		// simulate a response populating the cache after a miss
		flow := &px.Flow{
			Request: &px.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/test",
				},
				Header: http.Header{
					"Host":    []string{"example.com"},
					"Header1": []string{"value1"},
					"Header2": []string{"value2"},
				},
				Body: []byte("req"),
			},
		}
		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/plain"},
				"Header1":      []string{"value1"},
				"Header2":      []string{"value2"},
			},
			Body: []byte("resp"),
		}

		identifier := flow.Request.URL.String()
		len, err := respCacheAddon.cache.Len(identifier)
		require.Error(t, err, "error expected when checking length of non-existent bucket")
		require.Zero(t, len, "nothing in cache yet")

		// create traffic objects for the request and response, check header loading
		tReq, err := schema.NewProxyRequestFromMITMRequest(flow.Request, filterReqHeaders)
		require.NoError(t, err)
		require.Empty(t, tReq.Header.Get(CacheStatusHeader))
		require.Empty(t, tReq.Header.Get("header1"), "header should be deleted by factory function")
		require.NotEmpty(t, tReq.Header.Get("header2"), "header shouldn't be deleted by factory function")

		tResp, err := schema.NewProxyResponseFromMITMResponse(resp, filterRespHeaders)
		require.NoError(t, err)
		require.Empty(t, tResp.Header.Get(CacheStatusHeader))
		require.NotEmpty(t, tResp.Header.Get("header1"), "header should be deleted by factory function")
		require.Empty(t, tResp.Header.Get("header2"), "header shouldn't be deleted by factory function")

		// store the response in cache using an internal method, to simulate the real response storage
		respCacheAddon.cache.Put(tReq, tResp)

		// check length again, should work now
		len, err = respCacheAddon.cache.Len(identifier)
		require.NoError(t, err)
		require.Equal(t, 1, len)

		// simulate a new request with the same URL, should be a hit now that it's in the cache
		require.Empty(t, resp.Header.Get(CacheStatusHeader))
		respCacheAddon.Request(flow)
		require.NotNil(t, flow.Response)
		assert.Equal(t, resp.StatusCode, flow.Response.StatusCode)
		assert.Equal(t, resp.Body, flow.Response.Body)
		assert.Equal(t, "HIT", flow.Response.Header.Get(CacheStatusHeader))
	})

	t.Run("cache hit with gzip accept-encoding", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/test",
				},
				Header: http.Header{
					"Host":            []string{"example.com"},
					"Header1":         []string{"value1"},
					"Header2":         []string{"value2"},
					"Accept-Encoding": []string{"gzip"},
				},
				Body: []byte("req2"),
			},
		}
		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type":     []string{"text/plain"},
				"Header1":          []string{"value1"},
				"Header2":          []string{"value2"},
				"Content-Encoding": []string{"gzip"},
			},
		}
		// compress the body to simulate a real response
		respString := "resp2"
		encodedBody, _, err := utils.EncodeBody(&respString, "gzip")
		require.NoError(t, err)
		resp.Body = encodedBody

		// create traffic objects for the request and response, check header loading
		tReq, err := schema.NewProxyRequestFromMITMRequest(flow.Request, filterReqHeaders)
		require.NoError(t, err)
		require.Empty(t, tReq.Header.Get(CacheStatusHeader))
		require.Empty(t, tReq.Header.Get("header1"), "header should be deleted by factory function")
		require.NotEmpty(t, tReq.Header.Get("header2"), "header shouldn't be deleted by factory function")

		tResp, err := schema.NewProxyResponseFromMITMResponse(resp, filterRespHeaders)
		require.NoError(t, err)
		require.Empty(t, tResp.Header.Get(CacheStatusHeader))
		require.NotEmpty(t, tResp.Header.Get("header1"), "header should be deleted by factory function")
		require.Empty(t, tResp.Header.Get("header2"), "header shouldn't be deleted by factory function")
		require.Equal(t, "gzip", tResp.Header.Get("Content-Encoding"))

		// store the response in cache using an internal method, to simulate the real response storage
		respCacheAddon.cache.Put(tReq, tResp)

		// simulate a new request with the same URL, should be a hit now that it's in the cache
		require.Empty(t, resp.Header.Get(CacheStatusHeader))
		respCacheAddon.Request(flow)
		require.NotNil(t, flow.Response)
		assert.Equal(t, resp.StatusCode, flow.Response.StatusCode)
		assert.Equal(t, resp.Body, flow.Response.Body, fmt.Sprintf("expected resp.Body=%s to match flow.Response.Body=%s", string(resp.Body), string(flow.Response.Body)))
		assert.Equal(t, "HIT", flow.Response.Header.Get(CacheStatusHeader))
	})

	t.Cleanup(func() {
		respCacheAddon.Close()
	})
}
