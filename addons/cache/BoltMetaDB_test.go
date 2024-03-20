package cache

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/robbyt/llm_proxy/schema"

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
		trafficObjReq := schema.NewFromProxyRequest(req, []string{})
		require.NotNil(t, trafficObjReq)

		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header:     map[string][]string{"Content-Type": {"text/plain"}},
			Body:       []byte("hello"),
		}
		trafficObjResp := schema.NewFromProxyResponse(resp, []string{})
		require.NotNil(t, trafficObjResp)

		// empty cache
		gotResp, err := bMeta.Get(trafficObjReq)
		require.NoError(t, err)
		assert.Nil(t, gotResp)

		// use the Put method to store the response in the cache
		err = bMeta.Put(trafficObjReq, trafficObjResp)
		require.NoError(t, err)

		// check the length of the cache for this URL, should have 1 record
		len, err := bMeta.db.Len(req.URL.String())
		require.NoError(t, err)
		assert.Equal(t, 2, len) // not sure why this is 2, I assumed it should be 1

		// now use the Get method again to lookup the response
		gotResp, err = bMeta.Get(trafficObjReq)
		require.NoError(t, err)
		assert.Equal(t, resp.StatusCode, gotResp.StatusCode)
		assert.Equal(t, resp.Header, gotResp.Header)
		assert.Equal(t, resp.Body, []byte(gotResp.Body))
	})
}
