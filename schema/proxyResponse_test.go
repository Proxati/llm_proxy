package schema_test

import (
	"net/http"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementation of ProxyResponseReaderAdapter
type MockProxyResponseReaderAdapter struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (m *MockProxyResponseReaderAdapter) GetStatusCode() int {
	return m.StatusCode
}

func (m *MockProxyResponseReaderAdapter) GetHeaders() http.Header {
	return m.Headers
}

func (m *MockProxyResponseReaderAdapter) GetBodyBytes() []byte {
	return m.Body
}

func TestNewProxyResponseFromMITMResponse(t *testing.T) {
	// Test with nil input
	_, err := schema.NewProxyResponse(nil, nil)
	assert.Error(t, err)

	// Test with valid input
	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	headers.Add("Delete-Me", "too-many-secrets")
	mockAdapter := &MockProxyResponseReaderAdapter{
		StatusCode: 200,
		Headers:    headers,
		Body:       []byte(`{"key":"value"}`),
	}
	headersToFilter := []string{"Delete-Me"}

	res, err := schema.NewProxyResponse(mockAdapter, headersToFilter)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Status)
	assert.Contains(t, res.Header, "Content-Type")
	assert.NotContains(t, res.Header, "Delete-Me")
}
