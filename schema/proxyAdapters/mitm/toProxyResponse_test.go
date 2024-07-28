package mitm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockResponseReaderAdapter is a mock implementation of the ResponseReaderAdapter interface
type MockResponseReaderAdapter struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (m *MockResponseReaderAdapter) GetStatusCode() int {
	return m.StatusCode
}

func (m *MockResponseReaderAdapter) GetHeaders() http.Header {
	return m.Headers
}

func (m *MockResponseReaderAdapter) GetBodyBytes() []byte {
	return m.Body
}

func TestToProxyResponse(t *testing.T) {
	mockResponseReader := &MockResponseReaderAdapter{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": {"application/json"}},
		Body:       []byte(`{"key":"value"}`),
	}

	t.Run("Successful conversion", func(t *testing.T) {
		acceptEncodingHeader := "gzip"
		encodedBody, encoding, err := utils.EncodeBody(
			mockResponseReader.GetBodyBytes(), acceptEncodingHeader)
		require.NoError(t, err)

		resp, err := ToProxyResponse(mockResponseReader, acceptEncodingHeader)
		require.NoError(t, err)
		assert.Equal(t, mockResponseReader.GetStatusCode(), resp.StatusCode)

		// a few headers added by the encoding process
		assert.Equal(t, http.Header{
			"Content-Type":     {"application/json"},
			"Content-Encoding": {encoding},
			"Content-Length":   {fmt.Sprintf("%d", len(encodedBody))},
		}, resp.Header)
		assert.Equal(t, encodedBody, resp.Body)
	})

	t.Run("Invalid encoding means uncompressed response", func(t *testing.T) {
		acceptEncodingHeader := "invalid-encoding"
		encodedBody, encoding, err := utils.EncodeBody(
			mockResponseReader.GetBodyBytes(), acceptEncodingHeader)
		require.NoError(t, err)
		assert.Equal(t, []byte(`{"key":"value"}`), encodedBody)
		assert.Equal(t, "", encoding)

		resp, err := ToProxyResponse(mockResponseReader, acceptEncodingHeader)
		require.NoError(t, err)
		assert.Equal(t, mockResponseReader.GetStatusCode(), resp.StatusCode)

		// a few headers added by the encoding process
		assert.Equal(t, http.Header{
			"Content-Type":     {"application/json"},
			"Content-Encoding": {encoding},
			"Content-Length":   {fmt.Sprintf("%d", len(encodedBody))},
		}, resp.Header)
		assert.Equal(t, encodedBody, resp.Body)
	})
}
