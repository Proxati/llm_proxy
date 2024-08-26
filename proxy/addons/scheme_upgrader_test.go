package addons

import (
	"log/slog"
	"testing"

	"github.com/google/uuid"
	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestSchemeUpgrader_Request(t *testing.T) {
	upgrader := NewSchemeUpgrader(slog.Default())
	req := &px.Request{}
	err := req.UnmarshalJSON(
		[]byte(`{"method":"GET","url":"http://example.com","proto":"HTTP/1.1","header":{}}`))
	assert.Nil(t, err)

	flow := &px.Flow{
		Request: req,
		ConnContext: &px.ConnContext{
			ClientConn: &px.ClientConn{
				ID: uuid.UUID{},
			},
		},
	}

	upgrader.Request(flow)
	assert.Equal(t, "https", flow.Request.URL.Scheme)
	assert.Equal(t, "true", flow.Request.Header.Get(upgradedHeader))
}

func TestSchemeUpgrader_Request_HTTPS(t *testing.T) {
	upgrader := NewSchemeUpgrader(slog.Default())
	req := &px.Request{}
	err := req.UnmarshalJSON(
		[]byte(`{"method":"GET","url":"https://example.com","proto":"HTTP/1.1","header":{}}`))
	assert.Nil(t, err)

	flow := &px.Flow{
		Request: req,
		ConnContext: &px.ConnContext{
			ClientConn: &px.ClientConn{
				ID: uuid.UUID{},
			},
		},
	}

	upgrader.Request(flow)
	assert.Equal(t, "https", flow.Request.URL.Scheme)
	assert.Equal(t, "", flow.Request.Header.Get(upgradedHeader))
}
