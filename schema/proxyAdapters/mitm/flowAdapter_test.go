package mitm

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestFlowAdapterMiTM(t *testing.T) {
	pxFlow := &px.Flow{
		Request: &px.Request{
			Method: "GET",
			URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/flow"},
			Proto:  "HTTP/1.1",
			Header: http.Header{"User-Agent": []string{"TestAgent"}},
			Body:   []byte(`{"flow":"data"}`),
		},
		Response: &px.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"response":"data"}`),
		},
	}

	flowAdapter := NewFlowAdapter(pxFlow)

	assert.NotNil(t, flowAdapter)
	assert.Equal(t, pxFlow, flowAdapter.f)

	assert.Equal(t, "GET", flowAdapter.GetRequest().GetMethod())
	assert.Equal(t, "http://example.com/flow", flowAdapter.GetRequest().GetURL().String())
	assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
	assert.Equal(t, http.Header{"User-Agent": []string{"TestAgent"}}, flowAdapter.GetRequest().GetHeaders())
	assert.Equal(t, []byte(`{"flow":"data"}`), flowAdapter.GetRequest().GetBodyBytes())

	assert.Equal(t, 200, flowAdapter.GetResponse().GetStatusCode())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, flowAdapter.GetResponse().GetHeaders())
	assert.Equal(t, []byte(`{"response":"data"}`), flowAdapter.GetResponse().GetBodyBytes())
}
