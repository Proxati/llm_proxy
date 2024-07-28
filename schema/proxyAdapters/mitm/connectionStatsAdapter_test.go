package mitm

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestConnectionStatsAdapter(t *testing.T) {
	pxFlow := &px.Flow{
		ConnContext: &px.ConnContext{
			ClientConn: &px.ClientConn{},
		},
		Request: &px.Request{
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		},
		Id: uuid.New(),
	}

	statsAdapter := NewProxyConnectionStatsAdapter(pxFlow)

	assert.NotNil(t, statsAdapter)
	assert.Equal(t, pxFlow, statsAdapter.f)

	assert.Equal(t, UnknownAddr, statsAdapter.GetClientIP())
	assert.Equal(t, pxFlow.Id.String(), statsAdapter.GetProxyID())
	assert.Equal(t, "http://example.com", statsAdapter.GetRequestURL())
}
