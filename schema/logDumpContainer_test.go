package schema_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
)

type MockFlow struct {
	Request         *MockProxyRequestReaderAdapter
	Response        *MockProxyResponseReaderAdapter
	ConnectionStats *MockConnectionStatsReaderAdapter
	// ConnectionStats *Mock
	Id uuid.UUID
}

func (m MockFlow) GetRequest() proxyadapters.RequestReaderAdapter {
	return m.Request
}

func (m MockFlow) GetResponse() proxyadapters.ResponseReaderAdapter {
	return m.Response
}

func (m MockFlow) GetConnectionStats() proxyadapters.ConnectionStatsReaderAdapter {
	return m.ConnectionStats
}

func getDefaultFlow(t *testing.T) proxyadapters.FlowReaderAdapter {
	t.Helper()
	req := &MockProxyRequestReaderAdapter{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
		Proto: "HTTP/1.1",
		Headers: http.Header{
			"Content-Type":      []string{"[application/json]"},
			"Delete-Me-Request": []string{"too-many-secrets"},
		},
		Body: []byte(`{"key": "value"}`),
	}

	resp := &MockProxyResponseReaderAdapter{
		StatusCode: http.StatusOK,
		Headers: http.Header{
			"Content-Type":       []string{"[application/json]"},
			"Delete-Me-Response": []string{"too-many-secrets"},
		},
		Body: []byte(`{"status": "success"}`),
	}

	cs := &MockConnectionStatsReaderAdapter{}

	return MockFlow{
		Request:         req,
		Response:        resp,
		ConnectionStats: cs,
	}
}

func getDefaultConnectionStats(t *testing.T) *schema.ProxyConnectionStats {
	t.Helper()
	cs := &MockConnectionStatsReaderAdapter{}

	return &schema.ProxyConnectionStats{
		ClientAddress: cs.GetClientIP(),
		URL:           cs.GetRequestURL(),
		Duration:      0,
		ProxyID:       cs.GetProxyID(),
	}
}

func parseTimeStampString(t *testing.T, ts string) time.Time {
	t.Helper()
	expectedTimestamp, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatalf("failed to parse expected test timestamp: %v", err)
	}
	return expectedTimestamp
}

func TestNewLogDumpDiskContainer_JSON(t *testing.T) {
	t.Parallel()
	emptyHeaderFilterGroup := config.NewHeaderFilterGroup("empty", []string{}, []string{})
	basicHeaderFilterGroupReq := config.NewHeaderFilterGroup("basic-req", []string{}, []string{"Delete-Me-Request"})
	basicHeaderFilterGroupResp := config.NewHeaderFilterGroup("basic-req", []string{}, []string{"Delete-Me-Response"})

	testCases := []struct {
		name                    string
		flow                    proxyadapters.FlowReaderAdapter
		logSources              config.LogSourceConfig
		filterReqHeaders        *config.HeaderFilterGroup
		filterRespHeaders       *config.HeaderFilterGroup
		expectedConnectionStats *schema.ProxyConnectionStats
		expectedRequestMethod   string
		expectedRequestURL      string
		expectedRequestProto    string
		expectedRequestHeaders  string
		expectedRequestBody     string
		expectedResponseCode    int
		expectedResponseHeaders string
		expectedResponseBody    string
	}{
		{
			name: "all fields enabled",
			flow: getDefaultFlow(t),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        emptyHeaderFilterGroup,
			filterRespHeaders:       emptyHeaderFilterGroup,
			expectedConnectionStats: getDefaultConnectionStats(t),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\nDelete-Me-Request: too-many-secrets\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "Content-Type: [application/json]\r\nDelete-Me-Response: too-many-secrets\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields disabled",
			flow: getDefaultFlow(t),
			logSources: config.LogSourceConfig{
				LogConnectionStats: false,
				LogRequestHeaders:  false,
				LogRequest:         false,
				LogResponseHeaders: false,
				LogResponse:        false,
			},
			filterReqHeaders:        emptyHeaderFilterGroup,
			filterRespHeaders:       emptyHeaderFilterGroup,
			expectedConnectionStats: (*schema.ProxyConnectionStats)(nil), // weird way to assert nil
		},
		{
			name: "all fields enabled, with filter",
			flow: getDefaultFlow(t),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(t),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, request headers disabled",
			flow: getDefaultFlow(t),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  false,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(t),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, request disabled",
			flow: getDefaultFlow(t),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         false,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(t),
			expectedRequestHeaders:  "",
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseCode:    http.StatusOK,
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, response headers disabled",
			flow: getDefaultFlow(t),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: false,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(t),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, response disabled",
			flow: getDefaultFlow(t),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        false,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(t),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseHeaders: "",
			expectedResponseBody:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			container, err := schema.NewLogDumpContainer(tc.flow, tc.logSources, 0, tc.filterReqHeaders, tc.filterRespHeaders)
			require.Nil(t, err)
			assert.Equal(t, tc.expectedConnectionStats, container.ConnectionStats)
			assert.Equal(t, tc.expectedRequestMethod, container.Request.Method)
			assert.Equal(t, tc.expectedRequestURL, container.Request.URL.String())
			assert.Equal(t, tc.expectedRequestProto, container.Request.Proto)
			assert.Equal(t, tc.expectedRequestHeaders, container.Request.HeaderString())
			assert.Equal(t, tc.expectedRequestBody, container.Request.Body)
			assert.Equal(t, tc.expectedResponseCode, container.Response.Status)
			assert.Equal(t, tc.expectedResponseHeaders, container.Response.HeaderString())
			assert.Equal(t, tc.expectedResponseBody, container.Response.Body)
		})
	}
}

func TestUnmarshalLogDumpContainer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		jsonInput     string
		expectedError string
		expectedLDC   *schema.LogDumpContainer
	}{
		{
			name: "valid JSON with all fields",
			jsonInput: `
{
  "object_type": "llm_proxy_traffic_log",
  "schema": "v2",
  "timestamp": "2024-08-30T15:21:02.486594-04:00",
  "connection_stats": {
    "client_address": "127.0.0.1:49871",
    "url": "https://api.openai.com/v1/chat/completions",
    "duration_ms": 864,
    "proxy_id": "5d6f0016-d276-49ed-bb1d-788231477eae"
  },
  "request": {
    "url": "https://api.openai.com/v1/chat/completions",
    "method": "POST",
    "proto": "HTTP/1.1",
    "header": {
      "Content-Length": [
        "98"
      ],
      "Content-Type": [
        "application/json"
      ],
      "Proxy-Connection": [
        "Keep-Alive"
      ]
    },
    "body": "{\"messages\": [{\"role\": \"user\", \"content\": \"Hello, you are amazing?????\"}], \"model\": \"gpt-4o-mini\"}"
  },
  "response": {
    "status": 200,
    "header": {
	  "Content-Type": [
		"application/json"
	],
      "X-Request-Id": [
        "req_6f4313a0eed9f18da393ae94979bf6ed"
      ]
    },
    "body": "{\n  \"id\": \"chatcmpl-A21T4t9VS4OTmUDC2dEuA7X9TuwrQ\",\n  \"object\": \"chat.completion\",\n  \"created\": 1725045662,\n  \"model\": \"gpt-4o-mini-2024-07-18\",\n  \"choices\": [\n    {\n      \"index\": 0,\n      \"message\": {\n        \"role\": \"assistant\",\n        \"content\": \"Thank you! I appreciate the kind words. How can I assist you today?\",\n        \"refusal\": null\n      },\n      \"logprobs\": null,\n      \"finish_reason\": \"stop\"\n    }\n  ],\n  \"usage\": {\n    \"prompt_tokens\": 14,\n    \"completion_tokens\": 16,\n    \"total_tokens\": 30\n  },\n  \"system_fingerprint\": \"fp_f33667828e\"\n}\n"
  }
}`,
			expectedError: "",
			expectedLDC: &schema.LogDumpContainer{
				ObjectType:    schema.ObjectTypeDefault,
				SchemaVersion: schema.SchemaVersionV2,
				Timestamp:     parseTimeStampString(t, "2024-08-30T15:21:02.486594-04:00"),
				ConnectionStats: &schema.ProxyConnectionStats{
					ClientAddress: "127.0.0.1:49871",
					URL:           "https://api.openai.com/v1/chat/completions",
					Duration:      864,
					ProxyID:       "5d6f0016-d276-49ed-bb1d-788231477eae",
				},
				Request: &schema.ProxyRequest{
					Method: "POST",
					URL: &url.URL{
						Scheme:  "https",
						Opaque:  "",
						Host:    "api.openai.com",
						Path:    "/v1/chat/completions",
						RawPath: "",
					},
					Proto: "HTTP/1.1",
					Header: http.Header{
						"Content-Length":   []string{"98"},
						"Content-Type":     []string{"application/json"},
						"Proxy-Connection": []string{"Keep-Alive"},
					},
					Body: "{\"messages\": [{\"role\": \"user\", \"content\": \"Hello, you are amazing?????\"}], \"model\": \"gpt-4o-mini\"}",
				},
				Response: &schema.ProxyResponse{
					Status: 200,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
						"X-Request-Id": []string{"req_6f4313a0eed9f18da393ae94979bf6ed"},
					},
					Body: "{\n  \"id\": \"chatcmpl-A21T4t9VS4OTmUDC2dEuA7X9TuwrQ\",\n  \"object\": \"chat.completion\",\n  \"created\": 1725045662,\n  \"model\": \"gpt-4o-mini-2024-07-18\",\n  \"choices\": [\n    {\n      \"index\": 0,\n      \"message\": {\n        \"role\": \"assistant\",\n        \"content\": \"Thank you! I appreciate the kind words. How can I assist you today?\",\n        \"refusal\": null\n      },\n      \"logprobs\": null,\n      \"finish_reason\": \"stop\"\n    }\n  ],\n  \"usage\": {\n    \"prompt_tokens\": 14,\n    \"completion_tokens\": 16,\n    \"total_tokens\": 30\n  },\n  \"system_fingerprint\": \"fp_f33667828e\"\n}\n",
				},
			},
		},
		{
			name: "default schema version",
			jsonInput: `{
			    "object_type": "llm_proxy_traffic_log"
            }`,
			expectedError: "",
			expectedLDC: &schema.LogDumpContainer{
				ObjectType:    schema.ObjectTypeDefault,
				SchemaVersion: schema.DefaultSchemaVersion,
			},
		},
		{
			name: "invalid JSON missing object_type",
			jsonInput: `{
                "schema": "v2",
                "timestamp": "2023-10-01T12:00:00Z"
            }`,
			expectedError: "object_type is required",
			expectedLDC:   nil,
		},
		{
			name: "invalid JSON unsupported schema version",
			jsonInput: `{
                "object_type": "llm_proxy_traffic_log",
                "schema": "v1",
                "timestamp": "2023-10-01T12:00:00Z"
            }`,
			expectedError: "unsupported schema version",
			expectedLDC:   nil,
		},
		{
			name: "invalid schema type",
			jsonInput: `{
                "object_type": "llm_proxy_traffic_log",
                "schema": 1
            }`,
			expectedError: "schema must be a string",
			expectedLDC:   nil,
		},
		{
			name: "invalid object type",
			jsonInput: `{
                "object_type": 1,
                "schema": "v1"
            }`,
			expectedError: "object_type must be a string",
			expectedLDC:   nil,
		},
		{
			name: "invalid object type",
			jsonInput: `{
                "object_type": "foo",
                "schema": "v1"
            }`,
			expectedError: "unsupported object_type: foo",
			expectedLDC:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ldc, err := schema.UnmarshalLogDumpContainer([]byte(tc.jsonInput))
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLDC, ldc)
			}
		})
	}
}

func TestLogDumpContainer_MarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *schema.LogDumpContainer
		expected string
	}{
		{
			name: "all fields set",
			input: &schema.LogDumpContainer{
				ObjectType:    schema.ObjectTypeDefault,
				SchemaVersion: schema.DefaultSchemaVersion,
				Timestamp:     parseTimeStampString(t, "2023-01-01T00:00:00Z"),
				ConnectionStats: &schema.ProxyConnectionStats{
					ClientAddress: "127.0.0.1",
					URL:           "http://example.com",
					Duration:      100,
					ProxyID:       "proxy-1",
				},
				Request: &schema.ProxyRequest{
					Method: "GET",
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
					Proto: "HTTP/1.1",
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: `{"key": "value"}`,
				},
				Response: &schema.ProxyResponse{
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: `{"status": "success"}`,
				},
			},
			expected: `{"object_type":"llm_proxy_traffic_log","schema":"v2","timestamp":"2023-01-01T00:00:00Z","connection_stats":{"client_address":"127.0.0.1","url":"http://example.com","duration_ms":100,"proxy_id":"proxy-1"},"request":{"method":"GET","url":"http://example.com/","proto":"HTTP/1.1","header":{"Content-Type":["application/json"]},"body":"{\"key\": \"value\"}"},"response":{"header":{"Content-Type":["application/json"]},"body":"{\"status\": \"success\"}"}}`,
		},
		{
			name: "only mandatory fields set",
			input: &schema.LogDumpContainer{
				ObjectType:    schema.ObjectTypeDefault,
				SchemaVersion: schema.DefaultSchemaVersion,
			},
			expected: `{"object_type":"llm_proxy_traffic_log","schema":"v2"}`,
		},
		{
			name: "timestamp not set",
			input: &schema.LogDumpContainer{
				ObjectType:    schema.ObjectTypeDefault,
				SchemaVersion: schema.DefaultSchemaVersion,
				ConnectionStats: &schema.ProxyConnectionStats{
					ClientAddress: "127.0.0.1",
					URL:           "http://example.com",
					Duration:      100,
					ProxyID:       "proxy-1",
				},
			},
			expected: `{"object_type":"llm_proxy_traffic_log","schema":"v2","connection_stats":{"client_address":"127.0.0.1","url":"http://example.com","duration_ms":100,"proxy_id":"proxy-1"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.input.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, string(actual))
		})
	}
}
