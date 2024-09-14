package mitm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResponseReaderAdapter is a mock implementation of the ResponseReaderAdapter interface
type mockResponseReaderAdapter struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (m mockResponseReaderAdapter) GetStatusCode() int {
	return m.StatusCode
}

func (m mockResponseReaderAdapter) GetHeaders() http.Header {
	return m.Headers
}

func (m mockResponseReaderAdapter) GetBodyBytes() []byte {
	return m.Body
}

func TestToProxyResponse(t *testing.T) {
	tests := []struct {
		name                 string
		pRes                 proxyadapters.ResponseReaderAdapter
		acceptEncodingHeader string
		expectedStatusCode   int
		expectedHeaders      http.Header
		expectedBody         []byte
		expectError          bool
	}{
		// Existing test cases
		{
			name: "Successful conversion with gzip encoding",
			pRes: mockResponseReaderAdapter{
				StatusCode: 200,
				Headers:    http.Header{"Content-Type": {"application/json"}},
				Body:       []byte(`{"key":"value"}`),
			},
			acceptEncodingHeader: "gzip",
			expectedStatusCode:   200,
			expectedHeaders:      nil, // Will be set in the test
			expectedBody:         nil, // Will be set in the test
			expectError:          false,
		},
		{
			name: "Invalid encoding means uncompressed response",
			pRes: mockResponseReaderAdapter{
				StatusCode: 200,
				Headers:    http.Header{"Content-Type": {"application/json"}},
				Body:       []byte(`{"key":"value"}`),
			},
			acceptEncodingHeader: "invalid-encoding",
			expectedStatusCode:   200,
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		// New test cases
		{
			name: "Response with 404 Not Found",
			pRes: mockResponseReaderAdapter{
				StatusCode: 404,
				Headers:    http.Header{"Content-Type": {"text/plain"}},
				Body:       []byte("Not Found"),
			},
			acceptEncodingHeader: "gzip",
			expectedStatusCode:   404,
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "Response with existing Content-Encoding header",
			pRes: mockResponseReaderAdapter{
				StatusCode: 200,
				Headers: http.Header{
					"Content-Type":     {"application/json"},
					"Content-Encoding": {"deflate"},
				},
				Body: []byte(`{"key":"value"}`),
			},
			acceptEncodingHeader: "gzip",
			expectedStatusCode:   200,
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "Response with large body",
			pRes: mockResponseReaderAdapter{
				StatusCode: 200,
				Headers:    http.Header{"Content-Type": {"application/octet-stream"}},
				Body:       largeBodyHelper(1 * 1024 * 1024), // 1MB body
			},
			acceptEncodingHeader: "gzip",
			expectedStatusCode:   200,
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "Response with empty body",
			pRes: mockResponseReaderAdapter{
				StatusCode: 204, // No Content
				Headers:    http.Header{"Content-Type": {"application/json"}},
				Body:       []byte{},
			},
			acceptEncodingHeader: "gzip",
			expectedStatusCode:   204,
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Prepare expected values if no error is expected
			if !tt.expectError {
				encodedBody, encoding, err := utils.EncodeBody(tt.pRes.GetBodyBytes(), tt.acceptEncodingHeader)
				require.NoError(t, err)

				// Clone the original headers
				expectedHeaders := tt.pRes.GetHeaders().Clone()

				// Set Content-Length header
				expectedHeaders.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

				// Set expectedBody and adjust headers based on encoding
				if encoding == "" {
					tt.expectedBody = tt.pRes.GetBodyBytes()
					expectedHeaders.Del("Content-Encoding")
				} else {
					tt.expectedBody = encodedBody
					expectedHeaders.Set("Content-Encoding", encoding)
				}

				tt.expectedHeaders = expectedHeaders
			}

			// Call the function under test
			resp, err := ToProxyResponse(tt.pRes, tt.acceptEncodingHeader)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
				assert.Equal(t, tt.expectedHeaders, resp.Header)
				assert.Equal(t, tt.expectedBody, resp.Body)
			}
		})
	}
}
