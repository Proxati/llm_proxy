package addons

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters/mitm"
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
	t.Parallel()
	testLogger := slog.Default()
	emptyHeaderFilterGroup := config.NewHeaderFilterGroup(t.Name(), []string{}, []string{})

	t.Run("empty storage engine", func(t *testing.T) {
		storageEngineName := ""
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, emptyHeaderFilterGroup, emptyHeaderFilterGroup)
		assert.Error(t, err, "Expected error for empty storage engine")
		assert.Nil(t, cache)
	})

	t.Run("unknown storage engine", func(t *testing.T) {
		storageEngineName := "unknown"
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, emptyHeaderFilterGroup, emptyHeaderFilterGroup)
		assert.Error(t, err, "Expected error for unknown storage engine")
		assert.Nil(t, cache)
	})

	t.Run("bolt storage engine with invalid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := "\\\\invalid\\path"
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, emptyHeaderFilterGroup, emptyHeaderFilterGroup)
		assert.Error(t, err, "Expected error for invalid cacheDir")
		assert.Nil(t, cache)
	})

	t.Run("bolt storage engine with valid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(testLogger, storageEngineName, cacheDir, emptyHeaderFilterGroup, emptyHeaderFilterGroup)
		assert.NoError(t, err, "Expected no error for valid cacheDir")
		assert.NotNil(t, cache)
		assert.Equal(t, "ResponseCacheAddon", cache.String())
	})
}

func TestRequest(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()
	filterReqHeaders := config.NewHeaderFilterGroup(t.Name()+"req", []string{}, []string{"Header1"})
	filterRespHeaders := config.NewHeaderFilterGroup(t.Name()+"resp", []string{}, []string{"Header2"})

	newAddon := func() *ResponseCacheAddon {
		t.Helper()
		tmpDir := t.TempDir()
		t.Log("TempDir: ", tmpDir)
		respCacheAddon, err := NewCacheAddon(
			testLogger,
			"bolt", tmpDir,
			filterReqHeaders,
			filterRespHeaders,
		)
		require.Nil(t, err, "No error creating cache addon")
		return respCacheAddon
	}

	t.Run("closed addon", func(t *testing.T) {
		respCacheAddon := newAddon()
		err := respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
			},
		}

		respCacheAddon.Request(flow)
		require.NotNil(t, flow.Response, "Response should not be nil")
		assert.Equal(t, http.StatusServiceUnavailable, flow.Response.StatusCode, "Expected status code 503")
		assert.Equal(t, "LLM_Proxy is not available", string(flow.Response.Body), "Expected response body to match")
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"), "Expected Content-Type header to be text/plain")
		assert.Equal(t, CacheStatusSkip, flow.Response.Header.Get(CacheStatusHeader), "Expected CacheStatusHeader to be SKIP")
		assert.Equal(t, "close", flow.Response.Header.Get("Connection"), "Expected Connection header to be close")

		err = respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")
	})

	t.Run("open addon - cache miss", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/test-miss"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.Request(flow)
		assert.Equal(t, "MISS", flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be MISS")

		err := respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")
	})

	t.Run("open addon - cache hit", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
		}
		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/plain"},
			},
			Body: []byte("resp"),
		}

		// Store the response in cache
		reqAdapter := mitm.NewProxyRequestAdapter(flow.Request)
		tReq, err := schema.NewProxyRequest(reqAdapter, filterReqHeaders)
		require.NoError(t, err)

		respAdapter := mitm.NewProxyResponseAdapter(resp)
		tResp, err := schema.NewProxyResponse(respAdapter, filterRespHeaders)
		require.NoError(t, err)

		respCacheAddon.cache.Put(tReq, tResp)

		// Simulate the request hitting the addon
		respCacheAddon.Request(flow)
		require.NotNil(t, flow.Response, "Response should not be nil")
		assert.Equal(t, http.StatusOK, flow.Response.StatusCode, "Expected status code to match cached response")
		assert.Equal(t, "HIT", flow.Response.Header.Get(CacheStatusHeader), "Expected cache status to be HIT")

		err = respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")
	})

	t.Run("unsupported method: delete", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodDelete,
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.Request(flow)
		assert.Equal(
			t, CacheStatusSkip, flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be SKIP")

		err := respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")
	})

	t.Run("Cache-Control: no-cache", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":          []string{"example.com"},
					"Cache-Control": []string{"no-cache"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.Request(flow)
		assert.Equal(t, CacheStatusSkip, flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be SKIP")

		err := respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")
	})

	t.Run("Cache-Control: no-store", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":          []string{"example.com"},
					"Cache-Control": []string{"no-store"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.Request(flow)
		assert.Equal(t, CacheStatusSkip, flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be SKIP")

		err := respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")
	})

	t.Run("gzip encoding: hit", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/test",
				},
				Header: http.Header{
					"Host":            []string{"example.com"},
					"Accept-Encoding": []string{"gzip"},
				},
				Body: []byte("req"),
			},
		}

		// use internal gzipEncode function to encode the response body
		encodedBody, _, err := utils.EncodeBody([]byte("resp"), "gzip")
		require.NoError(t, err)

		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type":     []string{"text/plain"},
				"Content-Encoding": []string{"gzip"},
			},
			Body: encodedBody,
		}

		// Store the response in cache
		reqAdapter := mitm.NewProxyRequestAdapter(flow.Request)
		tReq, err := schema.NewProxyRequest(reqAdapter, filterReqHeaders)
		require.NoError(t, err)

		respAdapter := mitm.NewProxyResponseAdapter(resp)
		tResp, err := schema.NewProxyResponse(respAdapter, filterRespHeaders)
		require.NoError(t, err)

		respCacheAddon.cache.Put(tReq, tResp)

		// Simulate the request hitting the addon
		respCacheAddon.Request(flow)
		require.NotNil(t, flow.Response, "Response should not be nil")
		assert.Equal(t, http.StatusOK, flow.Response.StatusCode, "Expected status code to match cached response")
		assert.Equal(t, "HIT", flow.Response.Header.Get(CacheStatusHeader), "Expected cache status to be HIT")
		assert.Equal(t, "gzip", flow.Response.Header.Get("Content-Encoding"), "Expected Content-Encoding to be gzip")

		err = respCacheAddon.Close()
		require.NoError(t, err, "Expected no error closing addon")
	})

	t.Run("gzip encoding, miss/hit", func(t *testing.T) {
		respCacheAddon := newAddon()

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
		respBytes := []byte("resp2")
		encodedBody, _, err := utils.EncodeBody(respBytes, "gzip")
		require.NoError(t, err)
		resp.Body = encodedBody

		// convert the request to a RequestAdapter
		reqAdapter := mitm.NewProxyRequestAdapter(flow.Request)
		require.NotNil(t, reqAdapter)

		// create traffic objects for the request and response, check header loading
		tReq, err := schema.NewProxyRequest(reqAdapter, filterReqHeaders)
		require.NoError(t, err)
		require.Empty(t, tReq.Header.Get(CacheStatusHeader))
		require.Empty(t, tReq.Header.Get("header1"), "header should be deleted by factory function")
		require.NotEmpty(t, tReq.Header.Get("header2"), "header shouldn't be deleted by factory function")

		respAdapter := mitm.NewProxyResponseAdapter(resp)
		require.NotNil(t, respAdapter)

		tResp, err := schema.NewProxyResponse(respAdapter, filterRespHeaders)
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
}

func TestRequestClosed(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()
	respCacheAddon := &ResponseCacheAddon{}

	t.Run("requestClosed sets correct response", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
			},
		}

		respCacheAddon.requestClosed(testLogger, flow)

		require.NotNil(t, flow.Response, "Response should not be nil")
		assert.Equal(t, http.StatusServiceUnavailable, flow.Response.StatusCode, "Expected status code 503")
		assert.Equal(t, "LLM_Proxy is not available", string(flow.Response.Body), "Expected response body to match")
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"), "Expected Content-Type header to be text/plain")
		assert.Equal(t, CacheStatusSkip, flow.Response.Header.Get(CacheStatusHeader), "Expected CacheStatusHeader to be SKIP")
		assert.Equal(t, "close", flow.Response.Header.Get("Connection"), "Expected Connection header to be close")
	})
}

func TestRequestOpen(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()
	tmpDir := t.TempDir()
	filterReqHeaders := config.NewHeaderFilterGroup(t.Name()+"req", []string{}, []string{"Header1"})
	filterRespHeaders := config.NewHeaderFilterGroup(t.Name()+"resp", []string{}, []string{"Header2"})

	respCacheAddon, err := NewCacheAddon(
		testLogger,
		"memory", tmpDir,
		filterReqHeaders,
		filterRespHeaders,
	)
	require.Nil(t, err, "No error creating cache addon")

	t.Run("no Cache-Control header", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.requestOpen(testLogger, flow)
		assert.Equal(t, "MISS", flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be MISS")
	})

	t.Run("Cache-Control: no-cache", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":          []string{"example.com"},
					"Cache-Control": []string{"no-cache"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.requestOpen(testLogger, flow)
		assert.Equal(t, CacheStatusSkip, flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be SKIP")
	})

	t.Run("unsupported method", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "PUT",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.requestOpen(testLogger, flow)
		assert.Equal(t, CacheStatusSkip, flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be SKIP")
	})

	t.Run("cache hit", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
		}
		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/plain"},
			},
			Body: []byte("resp"),
		}

		// Store the response in cache
		reqAdapter := mitm.NewProxyRequestAdapter(flow.Request)
		tReq, err := schema.NewProxyRequest(reqAdapter, filterReqHeaders)
		require.NoError(t, err)

		respAdapter := mitm.NewProxyResponseAdapter(resp)
		tResp, err := schema.NewProxyResponse(respAdapter, filterRespHeaders)
		require.NoError(t, err)

		respCacheAddon.cache.Put(tReq, tResp)

		// Simulate the request hitting the addon
		respCacheAddon.requestOpen(testLogger, flow)
		require.NotNil(t, flow.Response, "Response should not be nil")
		assert.Equal(t, http.StatusOK, flow.Response.StatusCode, "Expected status code to match cached response")
		assert.Equal(t, "HIT", flow.Response.Header.Get(CacheStatusHeader), "Expected cache status to be HIT")
	})

	t.Run("cache miss", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test-miss"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
		}

		respCacheAddon.requestOpen(testLogger, flow)
		assert.Equal(t, "MISS", flow.Request.Header.Get(CacheStatusHeader), "Expected cache status to be MISS")
	})
}

func TestResponseCommon(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()
	tmpDir := t.TempDir()
	filterReqHeaders := config.NewHeaderFilterGroup(t.Name()+"req", []string{}, []string{"Header1"})
	filterRespHeaders := config.NewHeaderFilterGroup(t.Name()+"resp", []string{}, []string{"Header2"})

	respCacheAddon, err := NewCacheAddon(
		testLogger,
		"memory", tmpDir,
		filterReqHeaders,
		filterRespHeaders,
	)
	require.Nil(t, err, "No error creating cache addon")

	t.Run("CacheStatusHeader is set to SKIP", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":            []string{"example.com"},
					CacheStatusHeader: []string{CacheStatusSkip},
				},
			},
			Response: &px.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
			},
		}

		err := respCacheAddon.responseCommon(flow)
		require.Error(t, err, "Expected error for CacheStatusHeader set to SKIP")
		assert.Equal(t, CacheStatusSkip, flow.Response.Header.Get(CacheStatusHeader), "Expected CacheStatusHeader to be SKIP")
	})

	t.Run("CacheStatusHeader is set to MISS", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":            []string{"example.com"},
					CacheStatusHeader: []string{CacheStatusMiss},
				},
			},
			Response: &px.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
			},
		}

		err := respCacheAddon.responseCommon(flow)
		require.NoError(t, err, "Expected no error for CacheStatusHeader set to MISS")
		assert.Equal(t, CacheStatusMiss, flow.Response.Header.Get(CacheStatusHeader), "Expected CacheStatusHeader to be MISS")
	})

	t.Run("Unsupported response status code", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
			},
			Response: &px.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
			},
		}

		err := respCacheAddon.responseCommon(flow)
		require.Error(t, err, "Expected error for unsupported response status code")
		assert.Equal(t, CacheStatusSkip, flow.Response.Header.Get(CacheStatusHeader), "Expected CacheStatusHeader to be SKIP")
	})

	t.Run("Supported response status code", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
			},
			Response: &px.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
			},
		}

		err := respCacheAddon.responseCommon(flow)
		require.NoError(t, err, "Expected no error for supported response status code")
		assert.NotEqual(t, CacheStatusSkip, flow.Response.Header.Get(CacheStatusHeader), "Expected CacheStatusHeader not to be SKIP")
	})
}

func TestResponseStorage(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()
	filterReqHeaders := config.NewHeaderFilterGroup(t.Name()+"req", []string{}, []string{"Header1"})
	filterRespHeaders := config.NewHeaderFilterGroup(t.Name()+"resp", []string{}, []string{"Header2"})

	newAddon := func() *ResponseCacheAddon {
		t.Helper()
		tmpDir := t.TempDir()
		// t.Log("TempDir: ", tmpDir)
		respCacheAddon, err := NewCacheAddon(
			testLogger,
			"memory", tmpDir,
			filterReqHeaders,
			filterRespHeaders,
		)
		require.Nil(t, err, "No error creating cache addon")
		return respCacheAddon
	}

	t.Run("successful storage", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Request: &px.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/test",
				},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Body: []byte("req"),
			},
			Response: &px.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type":  []string{"text/plain"},
					"Random-Header": []string{"random"},
				},
				Body: []byte("resp"),
			},
		}

		err := respCacheAddon.responseStorage(flow)
		assert.NoError(t, err, "Expected no error during response storage")

		// lookup the response in the cache
		resp, err := respCacheAddon.cache.Get(flow.Request.URL.String(), flow.Request.Body)
		require.NoError(t, err, "Expected no error getting response from cache")
		require.NotNil(t, resp, "Expected response to be in cache")
		assert.Equal(
			t, http.StatusOK, resp.GetStatusCode(),
			"Expected status code to match cached response")
		assert.Equal(
			t, "text/plain", resp.GetHeaders().Get("Content-Type"),
			"Expected Content-Type header to match cached response")
		assert.Equal(
			t, "random", resp.GetHeaders().Get("Random-Header"),
			"Expected Random-Header to match cached response")
		assert.Equal(
			t, "resp", string(resp.GetBodyBytes()),
			"Expected response body to match cached response")
	})

	t.Run("error during request conversion", func(t *testing.T) {
		respCacheAddon := newAddon()

		flow := &px.Flow{
			Response: &px.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
				Body: []byte("resp"),
			},
		}

		err := respCacheAddon.responseStorage(flow)
		assert.Error(t, err, "Expected error during request conversion")
	})
}
