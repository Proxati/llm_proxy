package helpers

import (
	"log/slog"
	"net/http"
	"net/url"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	px "github.com/proxati/mitmproxy/proxy"
)

func TestRequestClosed(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()

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

		RequestClosed(testLogger, flow)

		require.NotNil(t, flow.Response, "Response should not be nil")
		assert.Equal(t, http.StatusServiceUnavailable, flow.Response.StatusCode, "Expected status code 503")
		assert.Equal(t, "LLM_Proxy is not available", string(flow.Response.Body), "Expected response body to match")
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"), "Expected Content-Type header to be text/plain")
		assert.Equal(t, headers.CacheStatusValueSkip, flow.Response.Header.Get(headers.CacheStatusHeader), "Expected headers.CacheStatusHeader to be SKIP")
		assert.Equal(t, "close", flow.Response.Header.Get("Connection"), "Expected Connection header to be close")
	})
}
