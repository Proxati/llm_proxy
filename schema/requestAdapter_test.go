package schema

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestRequestAdapterMiTM(t *testing.T) {
	pxReq := &px.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "https", Host: "example.com", Path: "/test"},
		Proto:  "HTTP/2.0",
		Header: http.Header{"Authorization": []string{"Bearer token"}},
		Body:   []byte(`{"data":"test"}`),
	}

	reqAccessor := NewRequestAdapter_MiTM(pxReq)

	assert.NotNil(t, reqAccessor)
	assert.Equal(t, pxReq, reqAccessor.pxReq)

	assert.Equal(t, "POST", reqAccessor.GetMethod())
	assert.Equal(t, "https://example.com/test", reqAccessor.GetURL().String())
	assert.Equal(t, "HTTP/2.0", reqAccessor.GetProto())
	assert.Equal(t, http.Header{"Authorization": []string{"Bearer token"}}, reqAccessor.GetHeaders())
	assert.Equal(t, []byte(`{"data":"test"}`), reqAccessor.GetBodyBytes())
}
