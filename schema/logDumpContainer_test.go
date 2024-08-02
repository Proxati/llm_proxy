package schema_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters"
)

type MockFlow struct {
	Request         *MockProxyRequestReaderAdapter
	Response        *MockProxyResponseReaderAdapter
	ConnectionStats *MockConnectionStatsReaderAdapter
	// ConnectionStats *Mock
	Id uuid.UUID
}

func (m MockFlow) GetRequest() proxyAdapters.RequestReaderAdapter {
	return m.Request
}

func (m MockFlow) GetResponse() proxyAdapters.ResponseReaderAdapter {
	return m.Response
}

func (m MockFlow) GetConnectionStats() proxyAdapters.ConnectionStatsReaderAdapter {
	return m.ConnectionStats
}

func getDefaultFlow() proxyAdapters.FlowReaderAdapter {
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

func getDefaultConnectionStats() *schema.ProxyConnectionStats {
	cs := &MockConnectionStatsReaderAdapter{}

	return &schema.ProxyConnectionStats{
		ClientAddress: cs.GetClientIP(),
		URL:           cs.GetRequestURL(),
		Duration:      0,
		ProxyID:       cs.GetProxyID(),
	}
}

func TestNewLogDumpDiskContainer_JSON(t *testing.T) {
	t.Parallel()
	emptyHeaderFilterGroup := config.NewHeaderFilterGroup("empty", []string{})
	basicHeaderFilterGroupReq := config.NewHeaderFilterGroup("basic-req", []string{"Delete-Me-Request"})
	basicHeaderFilterGroupResp := config.NewHeaderFilterGroup("basic-req", []string{"Delete-Me-Response"})

	testCases := []struct {
		name                    string
		flow                    proxyAdapters.FlowReaderAdapter
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
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        emptyHeaderFilterGroup,
			filterRespHeaders:       emptyHeaderFilterGroup,
			expectedConnectionStats: getDefaultConnectionStats(),
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
			flow: getDefaultFlow(),
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
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(),
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
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  false,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(),
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
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         false,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestHeaders:  "",
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseCode:    http.StatusOK,
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, response headers disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: false,
				LogResponse:        true,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(),
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
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        false,
			},
			filterReqHeaders:        basicHeaderFilterGroupReq,
			filterRespHeaders:       basicHeaderFilterGroupResp,
			expectedConnectionStats: getDefaultConnectionStats(),
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
