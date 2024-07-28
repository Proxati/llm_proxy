package schema

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestAccessor(t *testing.T) {
	// Test case for mitmproxy request
	request := &px.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/"},
		Proto:  "HTTP/1.1",
		Header: http.Header{"Content-Type": []string{"application/json"}},
		Body:   []byte(`{"key":"value"}`),
	}

	reqAccessor := NewRequestAccessor_MiTM(request)
	assert.NotNil(t, reqAccessor)
}

func TestRequestAccessorMiTM(t *testing.T) {
	pxReq := &px.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "https", Host: "example.com", Path: "/test"},
		Proto:  "HTTP/2.0",
		Header: http.Header{"Authorization": []string{"Bearer token"}},
		Body:   []byte(`{"data":"test"}`),
	}

	accessor := NewRequestAccessor_MiTM(pxReq)

	assert.Equal(t, "POST", accessor.GetMethod())
	assert.Equal(t, "https://example.com/test", accessor.GetURL().String())
	assert.Equal(t, "HTTP/2.0", accessor.GetProto())
	assert.Equal(t, http.Header{"Authorization": []string{"Bearer token"}}, accessor.GetHeaders())
	assert.Equal(t, []byte(`{"data":"test"}`), accessor.GetBodyBytes())
}
