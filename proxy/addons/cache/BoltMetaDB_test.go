package cache

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema"
	px "github.com/proxati/mitmproxy/proxy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBoltMetaDB(t *testing.T) {
	t.Run("valid db file", func(t *testing.T) {
		dbFileDir := t.TempDir()
		bMeta, err := NewBoltMetaDB(dbFileDir, []string{})

		require.NoError(t, err)
		assert.Equal(t, dbFileDir, bMeta.dbFileDir)
		assert.NotNil(t, bMeta.db)
		assert.NoError(t, bMeta.Close())
	})
}

func TestBoltMetaDB_PutAndGet(t *testing.T) {
	t.Run("put and get a request and response", func(t *testing.T) {
		dbFileDir := t.TempDir()
		bMeta, err := NewBoltMetaDB(dbFileDir, []string{})
		require.NoError(t, err)
		defer bMeta.Close()

		req := &px.Request{
			Method: "GET",
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/test",
			},
		}
		// convert the request to a RequestAccessor
		reqAccessor := schema.NewRequestAdapter_MiTM(req)
		require.NotNil(t, reqAccessor)

		trafficObjReq, err := schema.NewProxyRequest(reqAccessor, []string{})
		require.NoError(t, err)
		require.NotNil(t, trafficObjReq)

		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header:     map[string][]string{"Content-Type": {"text/plain"}},
			Body:       []byte("hello"),
		}

		// convert the response to a ResponseAccessor
		respAccessor := schema.NewResponseAdapter_MiTM(resp)
		require.NotNil(t, respAccessor)

		trafficObjResp, err := schema.NewProxyResponse(respAccessor, []string{})
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
		assert.Equal(t, resp.Header, gotResp.Header)
		assert.Equal(t, resp.Body, []byte(gotResp.Body))
	})
}
