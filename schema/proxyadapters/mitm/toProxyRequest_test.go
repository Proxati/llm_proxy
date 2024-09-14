package mitm

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func urlHelper(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid URL %s: %v", urlStr, err))
	}
	return u
}

func largeBodyHelper(size int) []byte {
	return bytes.Repeat([]byte("a"), size)
}

// mockRequestReaderAdapter is a mock implementation of RequestReaderAdapter for testing purposes
type mockRequestReaderAdapter struct {
	method  string
	url     *url.URL
	proto   string
	headers http.Header
	body    []byte
}

func (m mockRequestReaderAdapter) GetMethod() string {
	return m.method
}

func (m mockRequestReaderAdapter) GetURL() *url.URL {
	return m.url
}

func (m mockRequestReaderAdapter) GetProto() string {
	return m.proto
}

func (m mockRequestReaderAdapter) GetHeaders() http.Header {
	return m.headers
}

func (m mockRequestReaderAdapter) GetBodyBytes() []byte {
	return m.body
}

func TestToProxyRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		pReq                 proxyadapters.RequestReaderAdapter
		acceptEncodingHeader string
		expectedMethod       string
		expectedURL          string
		expectedHeaders      http.Header
		expectedBody         []byte
		expectError          bool
	}{
		{
			name: "Valid request with gzip encoding",
			pReq: mockRequestReaderAdapter{
				method: "GET",
				url:    urlHelper("http://example.com"),
				headers: http.Header{
					"Content-Type": []string{"application/json"},
				},
				body: []byte(`{"key":"value"}`),
			},
			acceptEncodingHeader: "gzip",
			expectedMethod:       "GET",
			expectedURL:          "http://example.com",
			expectedHeaders:      nil, // Will be set in the test
			expectedBody:         nil, // Will be set in the test
			expectError:          false,
		},
		{
			name: "Invalid encoding means uncompressed request",
			pReq: mockRequestReaderAdapter{
				method: "POST",
				url:    urlHelper("http://example.com"),
				headers: http.Header{
					"Content-Type": []string{"application/json"},
				},
				body: []byte(`{"key":"value"}`),
			},
			acceptEncodingHeader: "invalid-encoding",
			expectedMethod:       "POST",
			expectedURL:          "http://example.com",
			expectedHeaders:      nil, // Will be set in the test
			expectedBody:         nil, // Will be set in the test
			expectError:          false,
		},
		{
			name: "PUT request with gzip encoding",
			pReq: mockRequestReaderAdapter{
				method: "PUT",
				url:    urlHelper("http://example.com/put"),
				headers: http.Header{
					"Content-Type": []string{"application/json"},
				},
				body: []byte(`{"update":"data"}`),
			},
			acceptEncodingHeader: "gzip",
			expectedMethod:       "PUT",
			expectedURL:          "http://example.com/put",
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "DELETE request with gzip encoding",
			pReq: mockRequestReaderAdapter{
				method: "DELETE",
				url:    urlHelper("http://example.com/delete"),
				headers: http.Header{
					"Content-Type": []string{"application/json"},
				},
				body: []byte(``), // Empty body
			},
			acceptEncodingHeader: "gzip",
			expectedMethod:       "DELETE",
			expectedURL:          "http://example.com/delete",
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "GET request with query parameters and gzip encoding",
			pReq: mockRequestReaderAdapter{
				method: "GET",
				url:    urlHelper("http://example.com/search?q=golang&sort=asc"),
				headers: http.Header{
					"Accept": []string{"application/json"},
				},
				body: []byte(``), // Empty body
			},
			acceptEncodingHeader: "gzip",
			expectedMethod:       "GET",
			expectedURL:          "http://example.com/search?q=golang&sort=asc",
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "POST request with existing Content-Encoding header",
			pReq: mockRequestReaderAdapter{
				method: "POST",
				url:    urlHelper("http://example.com/submit"),
				headers: http.Header{
					"Content-Type":     []string{"application/json"},
					"Content-Encoding": []string{"deflate"},
				},
				body: []byte(`{"key":"value"}`),
			},
			acceptEncodingHeader: "gzip",
			expectedMethod:       "POST",
			expectedURL:          "http://example.com/submit",
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "POST request with large body and gzip encoding",
			pReq: mockRequestReaderAdapter{
				method: "POST",
				url:    urlHelper("http://example.com/large"),
				headers: http.Header{
					"Content-Type": []string{"application/octet-stream"},
				},
				body: largeBodyHelper(1 * 1024 * 1024), // 1MB body
			},
			acceptEncodingHeader: "gzip",
			expectedMethod:       "POST",
			expectedURL:          "http://example.com/large",
			expectedHeaders:      nil,
			expectedBody:         nil,
			expectError:          false,
		},
		{
			name: "GET request with empty body and gzip encoding",
			pReq: mockRequestReaderAdapter{
				method: "GET",
				url:    urlHelper("http://example.com/empty"),
				headers: http.Header{
					"Accept": []string{"application/json"},
				},
				body: []byte{}, // Empty body
			},
			acceptEncodingHeader: "gzip",
			expectedMethod:       "GET",
			expectedURL:          "http://example.com/empty",
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
				encodedBody, encoding, err := utils.EncodeBody(tt.pReq.GetBodyBytes(), tt.acceptEncodingHeader)
				require.NoError(t, err)

				// Clone the original headers
				tt.expectedHeaders = tt.pReq.GetHeaders().Clone()

				// Set Content-Length header
				tt.expectedHeaders.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

				// Set expectedBody and adjust headers based on encoding
				if encoding == "" {
					tt.expectedBody = tt.pReq.GetBodyBytes()
					tt.expectedHeaders.Del("Content-Encoding")
				} else {
					tt.expectedBody = encodedBody
					tt.expectedHeaders.Set("Content-Encoding", encoding)
				}
			}

			// Call the function under test
			req, err := ToProxyRequest(tt.pReq, tt.acceptEncodingHeader)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMethod, req.Method)
				assert.Equal(t, tt.expectedURL, req.URL.String())
				assert.Equal(t, tt.expectedHeaders, req.Header)
				assert.Equal(t, tt.expectedBody, req.Body)
			}
		})
	}
}
