package schema

import (
	"net/http"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProxyResponseFromMITMResponse(t *testing.T) {
	// Test with nil input
	_, err := NewProxyResponseFromMITMResponse(nil, nil)
	assert.Error(t, err)

	// Test with valid input
	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	headers.Add("Delete-Me", "too-many-secrets")
	req := &px.Response{
		StatusCode: 200,
		Header:     headers,
		Body:       []byte(`{"key":"value"}`),
	}
	headersToFilter := []string{"Delete-Me"}

	res, err := NewProxyResponseFromMITMResponse(req, headersToFilter)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Status)
	assert.Contains(t, res.Header, "Content-Type")
	assert.NotContains(t, res.Header, "Delete-Me")
}
