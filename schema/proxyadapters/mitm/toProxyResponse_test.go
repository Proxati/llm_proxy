package mitm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResponseReaderAdapter is a mock implementation of the ResponseReaderAdapter interface
type mockResponseReaderAdapter struct {
	t          *testing.T
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (m mockResponseReaderAdapter) GetStatusCode() int {
	m.t.Helper()
	return m.StatusCode
}

func (m mockResponseReaderAdapter) GetHeaders() http.Header {
	m.t.Helper()
	return m.Headers
}

func (m mockResponseReaderAdapter) GetBodyBytes() []byte {
	m.t.Helper()
	return m.Body
}

func TestToProxyResponse(t *testing.T) {
	t.Parallel()
	mockResponseReader := &mockResponseReaderAdapter{
		t:          t,
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

		// Build expected headers
		expectedHeaders := http.Header{
			"Content-Type":     {"application/json"},
			"Content-Encoding": {encoding},
			"Content-Length":   {fmt.Sprintf("%d", len(encodedBody))},
		}

		// Compare headers
		assert.Equal(t, expectedHeaders, resp.Header)
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

		// Build expected headers
		expectedHeaders := http.Header{
			"Content-Type":   {"application/json"},
			"Content-Length": {fmt.Sprintf("%d", len(encodedBody))},
		}

		// Compare headers
		assert.Equal(t, expectedHeaders, resp.Header)
		assert.Equal(t, encodedBody, resp.Body)
	})
}
