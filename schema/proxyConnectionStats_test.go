package schema

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestLogStdOutLine_toJSONstr(t *testing.T) {
	line := &ProxyConnectionStats{
		ClientAddress: "127.0.0.1",
		URL:           "http://example.com",
		Duration:      100,
	}

	expected := `{"client_address":"127.0.0.1","url":"http://example.com","duration_ms":100}`
	assert.Equal(t, expected, line.toJSONstr())
}

func TestNewLogLine(t *testing.T) {
	// Create a mock Flow object
	f := px.Flow{
		Request: &px.Request{
			Method: "GET",
			URL:    &url.URL{Scheme: "https", Host: "example.com", Path: "/testpath"},
		},
		Response: &px.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}, "X-Request-Id": []string{"1234"}},
			Body:       nil, // You can also mock a Body here
		},
		Id: uuid.NewV4(),
	}

	csAdapter := newProxyConnectionStatsAdapter_MiTM(&f)
	logLine := newProxyConnectionStatsWithDuration(csAdapter, 100)
	assert.NotNil(t, logLine)
	assert.Equal(t, "unknown", logLine.ClientAddress)
	assert.Equal(t, "https://example.com/testpath", logLine.URL)
	assert.Equal(t, f.Id.String(), logLine.ProxyID)
	assert.Equal(t, int64(100), logLine.Duration)
}
