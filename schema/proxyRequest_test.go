package schema

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyStringSlice = []string{}

func Test_NewFromMITMRequest(t *testing.T) {
	t.Run("new from proxy request", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"application/json"},
			"Delete-Me":    []string{"too-many-secrets"},
		}
		headersToFilter := []string{"Delete-Me"}

		url, err := url.Parse("http://example.com")
		require.NoError(t, err)

		request := &px.Request{
			Method: "GET",
			URL:    url,
			Header: headers,
			Body:   []byte("hello"),
			Proto:  "HTTP/1.1",
		}

		// convert the request to a RequestAccessor
		reqAccessor, err := NewRequestAccessor(request)
		require.NoError(t, err)

		trafficObject, err := NewProxyRequest(reqAccessor, headersToFilter)
		require.NoError(t, err)
		assert.Equal(t, "GET", trafficObject.Method)
		assert.Equal(t, "http://example.com", trafficObject.URL.String())
		assert.Contains(t, trafficObject.Header, "Content-Type")
		assert.NotContains(t, trafficObject.Header, "Delete-Me")
		assert.Equal(t, "hello", trafficObject.Body)
		assert.Equal(t, "HTTP/1.1", trafficObject.Proto)
	})
	t.Run("new from proxy request with binary body", func(t *testing.T) {
		request := &px.Request{
			Body: []byte("\x01\x02\x03"),
		}
		// convert the request to a RequestAccessor
		reqAccessor, err := NewRequestAccessor(request)
		require.NoError(t, err)

		trafficObject, err := NewProxyRequest(reqAccessor, emptyStringSlice)
		require.NoError(t, err)
		assert.NotNil(t, trafficObject)
		assert.Empty(t, trafficObject.Body)
	})
	t.Run("nil request", func(t *testing.T) {
		trafficObject, err := NewProxyRequest(nil, emptyStringSlice)
		require.Error(t, err)
		assert.Nil(t, trafficObject)
	})
}

func TestProxyRequest_UnmarshalJSON(t *testing.T) {
	t.Run("successful unmarshal", func(t *testing.T) {
		data := []byte(`{
			"method": "GET",
			"url": "http://example.com",
			"header": {
				"Content-Type": ["application/json"]
			},
			"body": "hello",
			"proto": "HTTP/1.1"
		}`)
		pReq := &ProxyRequest{}
		err := pReq.UnmarshalJSON(data)
		require.NoError(t, err)
		assert.Equal(t, "GET", pReq.Method)
		assert.Equal(t, "http://example.com", pReq.URL.String())
		assert.Equal(t, []string{"application/json"}, pReq.Header["Content-Type"])
		assert.Equal(t, "hello", pReq.Body)
		assert.Equal(t, "HTTP/1.1", pReq.Proto)
	})

	t.Run("unmarshal with invalid url", func(t *testing.T) {
		data := []byte(`{
			"url": "://invalid_url"
		}`)
		pReq := &ProxyRequest{}
		err := pReq.UnmarshalJSON(data)
		require.Error(t, err)
	})

	t.Run("unmarshal with invalid headers", func(t *testing.T) {
		data := []byte(`{
			"header": {
				"Content-Type": "invalid_header"
			}
		}`)
		pReq := &ProxyRequest{}
		err := pReq.UnmarshalJSON(data)
		require.Error(t, err)
	})
}

func TestProxyRequest_MarshalJSON(t *testing.T) {
	t.Run("successful marshal", func(t *testing.T) {
		pReq := &ProxyRequest{
			Method: "GET",
			URL:    &url.URL{Scheme: "http", Host: "example.com"},
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:  "hello",
			Proto: "HTTP/1.1",
		}

		data, err := pReq.MarshalJSON()
		require.NoError(t, err)

		expected := `{"url":"http://example.com","method":"GET","header":{"Content-Type":["application/json"]},"body":"hello","proto":"HTTP/1.1"}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("marshal with nil URL", func(t *testing.T) {
		pReq := &ProxyRequest{
			Method: "GET",
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:  "hello",
			Proto: "HTTP/1.1",
		}

		data, err := pReq.MarshalJSON()
		require.NoError(t, err)

		expected := `{"method":"GET","header":{"Content-Type":["application/json"]},"body":"hello","proto":"HTTP/1.1"}`
		assert.JSONEq(t, expected, string(data))
	})
}
