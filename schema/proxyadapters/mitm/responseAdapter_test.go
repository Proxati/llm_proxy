package mitm

import (
	"net/http"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestResponseAdapterMiTM(t *testing.T) {
	pxResp := &px.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"key":"value"}`),
	}

	respAdapter := NewProxyResponseAdapter(pxResp)

	assert.NotNil(t, respAdapter)
	assert.Equal(t, pxResp, respAdapter.pxResp)

	assert.Equal(t, 200, respAdapter.GetStatusCode())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, respAdapter.GetHeaders())
	assert.Equal(t, []byte(`{"key":"value"}`), respAdapter.GetBodyBytes())
}
