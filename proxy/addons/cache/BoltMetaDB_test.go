package cache

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters/mitm"
	px "github.com/proxati/mitmproxy/proxy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBoltMetaDB(t *testing.T) {
	respHeaderFilter := config.NewHeaderFilterGroup([]string{})

	t.Run("valid db file", func(t *testing.T) {
		dbFileDir := t.TempDir()
		bMeta, err := NewBoltMetaDB(dbFileDir, respHeaderFilter)

		require.NoError(t, err)
		assert.Equal(t, dbFileDir, bMeta.dbFileDir)
		assert.NotNil(t, bMeta.db)
		assert.NoError(t, bMeta.Close())
	})
}

func TestBoltMetaDB_PutAndGet(t *testing.T) {
	reqHeaderFilter := config.NewHeaderFilterGroup([]string{})
	respHeaderFilter := config.NewHeaderFilterGroup([]string{"Set-Cookie"})

	t.Run("put and get a request and response", func(t *testing.T) {
		dbFileDir := t.TempDir()
		bMeta, err := NewBoltMetaDB(dbFileDir, respHeaderFilter)
		require.NoError(t, err)
		defer bMeta.Close()

		req := &px.Request{
			Method: "GET",
			Header: http.Header{},
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/test",
			},
		}
		// convert the request to a RequestAdapter
		reqAdapter := mitm.NewProxyRequestAdapter(req)
		require.NotNil(t, reqAdapter)

		trafficObjReq, err := schema.NewProxyRequest(reqAdapter, reqHeaderFilter)
		require.NoError(t, err)
		require.NotNil(t, trafficObjReq)
		require.Empty(t, trafficObjReq.Header)

		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Set-Cookie":   {"cookie1=123; cookie2=456"}, // this should be filtered out
				"Content-Type": {"text/plain"},               // this should be kept
			},
			Body: []byte("hello"),
		}

		// convert the response to a ResponseAdapter
		respAdapter := mitm.NewProxyResponseAdapter(resp)
		require.NotNil(t, respAdapter)

		trafficObjResp, err := schema.NewProxyResponse(respAdapter, respHeaderFilter)
		require.NoError(t, err)
		require.NotNil(t, trafficObjResp)

		// empty cache
		gotResp, err := bMeta.Get(trafficObjReq.URL.String(), []byte{})
		require.NoError(t, err)
		assert.Nil(t, gotResp)

		// use the Put method to store the response in the cache
		err = bMeta.Put(trafficObjReq, trafficObjResp)
		require.NoError(t, err)

		// check the length of the cache for this URL, should have 1 record
		len, err := bMeta.db.Len(req.URL.String())
		require.NoError(t, err)
		assert.Equal(t, 1, len)

		// now use the Get method again to lookup the response
		gotResp, err = bMeta.Get(trafficObjReq.URL.String(), []byte{})
		require.NoError(t, err)
		assert.Equal(t, resp.StatusCode, gotResp.Status)
		assert.Equal(t, resp.Body, []byte(gotResp.Body))

		// headers are filtered
		assert.NotEqual(t, resp.Header, gotResp.Header)
		assert.Equal(t, http.Header{"Content-Type": {"text/plain"}}, gotResp.Header)

	})
}
