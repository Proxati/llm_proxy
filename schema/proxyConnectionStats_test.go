package schema_test

import (
	"testing"

	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/stretchr/testify/assert"
)

// MockConnectionStatsReaderAdapter is a mock implementation of the proxyAdapters.ConnectionStatsReaderAdapter interface
type MockConnectionStatsReaderAdapter struct{}

func (m *MockConnectionStatsReaderAdapter) GetClientIP() string {
	return "192.168.1.1"
}

func (m *MockConnectionStatsReaderAdapter) GetProxyID() string {
	return "mock-proxy-id"
}

func (m *MockConnectionStatsReaderAdapter) GetRequestURL() string {
	return "http://mockurl.com"
}

func TestNewProxyConnectionStatsWithDuration(t *testing.T) {
	mockAdapter := &MockConnectionStatsReaderAdapter{}
	duration := int64(150)

	stats := schema.NewProxyConnectionStatsWithDuration(mockAdapter, duration)

	assert.NotNil(t, stats)
	assert.Equal(t, "192.168.1.1", stats.ClientAddress)
	assert.Equal(t, "mock-proxy-id", stats.ProxyID)
	assert.Equal(t, "http://mockurl.com", stats.URL)
	assert.Equal(t, duration, stats.Duration)
}

func TestLogStdOutLine_toJSONstr(t *testing.T) {
	line := &schema.ProxyConnectionStats{
		ClientAddress: "127.0.0.1",
		URL:           "http://example.com",
		Duration:      100,
	}

	expected := `{"client_address":"127.0.0.1","url":"http://example.com","duration_ms":100}`
	assert.Equal(t, expected, line.ToJSONstr())
}
