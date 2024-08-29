package mitm

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

	reqAdapter := NewProxyRequestAdapter(pxReq)

	assert.NotNil(t, reqAdapter)
	assert.Equal(t, pxReq, reqAdapter.pxReq)

	assert.Equal(t, "POST", reqAdapter.GetMethod())
	assert.Equal(t, "https://example.com/test", reqAdapter.GetURL().String())
	assert.Equal(t, "HTTP/2.0", reqAdapter.GetProto())
	assert.Equal(t, http.Header{"Authorization": []string{"Bearer token"}}, reqAdapter.GetHeaders())
	assert.Equal(t, []byte(`{"data":"test"}`), reqAdapter.GetBodyBytes())
}
