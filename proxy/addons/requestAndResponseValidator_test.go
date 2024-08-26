package addons

import (
	"log/slog"
	"net/http"
	"net/url"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestRequestAndResponseValidator_Request(t *testing.T) {
	logger := slog.Default()
	validator := NewRequestAndResponseValidator(logger)

	t.Run("Request is nil", func(t *testing.T) {
		flow := &px.Flow{
			Request:     nil,
			ConnContext: &px.ConnContext{},
		}
		assert.NotPanics(t, func() {
			validator.Request(flow)
		})
		assert.Equal(t, http.StatusBadRequest, flow.Response.StatusCode)
		assert.Equal(t, "Request is nil", string(flow.Response.Body))
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"))
	})

	t.Run("Request URL is nil", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				URL: nil,
			},
			ConnContext: &px.ConnContext{},
		}
		assert.NotPanics(t, func() {
			validator.Request(flow)
		})
		assert.Equal(t, http.StatusBadRequest, flow.Response.StatusCode)
		assert.Equal(t, "Request URL is empty", string(flow.Response.Body))
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"))
	})

	t.Run("Request URL is empty", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				URL: &url.URL{},
			},
			ConnContext: &px.ConnContext{},
		}
		assert.NotPanics(t, func() {
			validator.Request(flow)
		})
		assert.Equal(t, http.StatusBadRequest, flow.Response.StatusCode)
		assert.Equal(t, "Request URL is empty", string(flow.Response.Body))
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"))
	})

	t.Run("Valid Request", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				URL: &url.URL{Path: "/valid"},
			},
			ConnContext: &px.ConnContext{},
		}
		assert.NotPanics(t, func() {
			validator.Request(flow)
		})
		assert.Nil(t, flow.Response)
	})
}

func TestRequestAndResponseValidator_Response(t *testing.T) {
	logger := slog.Default()
	validator := NewRequestAndResponseValidator(logger)

	t.Run("Response is nil", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				URL: &url.URL{Path: "/valid"},
			},
			Response:    nil,
			ConnContext: &px.ConnContext{},
		}
		assert.NotPanics(t, func() {
			validator.Response(flow)
		})
		assert.Equal(t, http.StatusBadGateway, flow.Response.StatusCode)
		assert.Equal(t, "Response is nil", string(flow.Response.Body))
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"))
	})

	t.Run("Valid Response", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				URL: &url.URL{Path: "/valid"},
			},
			Response: &px.Response{
				StatusCode: http.StatusOK,
				Body:       []byte("OK"),
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
			},
			ConnContext: &px.ConnContext{},
		}
		assert.NotPanics(t, func() {
			validator.Response(flow)
		})
		assert.Equal(t, http.StatusOK, flow.Response.StatusCode)
		assert.Equal(t, "OK", string(flow.Response.Body))
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"))
	})
}
