package mitm

import (
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

func TestToProxyRequest(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare expected values if no error is expected
			if !tt.expectError {
				encodedBody, encoding, err := utils.EncodeBody(tt.pReq.GetBodyBytes(), tt.acceptEncodingHeader)
				require.NoError(t, err)
				tt.expectedBody = encodedBody

				// Clone the original headers
				tt.expectedHeaders = make(http.Header)
				for k, v := range tt.pReq.GetHeaders() {
					tt.expectedHeaders[k] = v
				}

				tt.expectedHeaders.Set("Content-Encoding", encoding)
				tt.expectedHeaders.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))
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
