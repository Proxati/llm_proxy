package schema

import (
	"net/http"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestResponseAccessorMiTM(t *testing.T) {
	pxResp := &px.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"key":"value"}`),
	}

	respAccessor := NewResponseAdapter_MiTM(pxResp)

	assert.NotNil(t, respAccessor)
	assert.Equal(t, pxResp, respAccessor.pxResp)

	assert.Equal(t, 200, respAccessor.GetStatusCode())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, respAccessor.GetHeaders())
	assert.Equal(t, []byte(`{"key":"value"}`), respAccessor.GetBodyBytes())
}
