package mitm

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/proxati/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestFlowAdapterMiTM(t *testing.T) {
	t.Parallel()
	pxFlow := &px.Flow{
		Request: &px.Request{
			Method: "GET",
			URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/flow"},
			Proto:  "HTTP/1.1",
			Header: http.Header{"User-Agent": []string{"TestAgent"}},
			Body:   []byte(`{"flow":"data"}`),
		},
		Response: &px.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"response":"data"}`),
		},
	}

	flowAdapter := NewFlowAdapter(pxFlow)

	assert.NotNil(t, flowAdapter)
	assert.Equal(t, pxFlow, flowAdapter.connectionStats.f)

	assert.Equal(t, "GET", flowAdapter.GetRequest().GetMethod())
	assert.Equal(t, "http://example.com/flow", flowAdapter.GetRequest().GetURL().String())
	assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
	assert.Equal(t, http.Header{"User-Agent": []string{"TestAgent"}}, flowAdapter.GetRequest().GetHeaders())
	assert.Equal(t, []byte(`{"flow":"data"}`), flowAdapter.GetRequest().GetBodyBytes())

	assert.Equal(t, 200, flowAdapter.GetResponse().GetStatusCode())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, flowAdapter.GetResponse().GetHeaders())
	assert.Equal(t, []byte(`{"response":"data"}`), flowAdapter.GetResponse().GetBodyBytes())
}

func TestFlowAdapterSetRequest(t *testing.T) {
	t.Parallel()
	t.Run("Typical", func(t *testing.T) {
		flowAdapter := &FlowAdapter{}
		req := &px.Request{
			Method: "POST",
			URL:    &url.URL{Scheme: "https", Host: "example.com", Path: "/newflow"},
			Proto:  "HTTP/2.0",
			Header: http.Header{"User-Agent": []string{"NewTestAgent"}},
			Body:   []byte(`{"newflow":"data"}`),
		}
		flowAdapter.SetRequest(req)

		assert.Equal(t, "POST", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "https://example.com/newflow", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "HTTP/2.0", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{"User-Agent": []string{"NewTestAgent"}}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(`{"newflow":"data"}`), flowAdapter.GetRequest().GetBodyBytes())
	})

	t.Run("NilRequest", func(t *testing.T) {
		flowAdapter := &FlowAdapter{}
		flowAdapter.SetRequest(nil)

		assert.Equal(t, "", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(nil), flowAdapter.GetRequest().GetBodyBytes())
	})

	t.Run("NilURL", func(t *testing.T) {
		flowAdapter := &FlowAdapter{}
		req := &px.Request{
			Method: "GET",
			URL:    nil, // should defend against NPEs from nil URL
			Proto:  "HTTP/1.1",
			Header: http.Header{"User-Agent": []string{"NilURLAgent"}},
			Body:   []byte(`{"nilurl":"data"}`),
		}
		flowAdapter.SetRequest(req)

		assert.Equal(t, "GET", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{"User-Agent": []string{"NilURLAgent"}}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(`{"nilurl":"data"}`), flowAdapter.GetRequest().GetBodyBytes())
	})

	t.Run("NilHeaders", func(t *testing.T) {
		flowAdapter := &FlowAdapter{}
		req := &px.Request{
			Method: "GET",
			URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/nilheaders"},
			Proto:  "HTTP/1.1",
			Header: nil, // should defend against NPEs from nil headers
			Body:   []byte(`{"nilheaders":"data"}`),
		}
		flowAdapter.SetRequest(req)

		assert.Equal(t, "GET", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "http://example.com/nilheaders", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(`{"nilheaders":"data"}`), flowAdapter.GetRequest().GetBodyBytes())
	})

	t.Run("SetRequestTwice", func(t *testing.T) {
		flowAdapter := &FlowAdapter{}
		req := &px.Request{
			Method: "GET",
			URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/twice"},
			Proto:  "HTTP/1.1",
			Header: http.Header{"User-Agent": []string{"TwiceAgent"}},
			Body:   []byte(`{"twice":"data"}`),
		}
		flowAdapter.SetRequest(req)

		assert.Equal(t, "GET", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "http://example.com/twice", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{"User-Agent": []string{"TwiceAgent"}}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(`{"twice":"data"}`), flowAdapter.GetRequest().GetBodyBytes())

		req2 := &px.Request{
			Method: "POST",
			URL:    &url.URL{Scheme: "https", Host: "example.com", Path: "/twice"},
			Proto:  "HTTP/2.0",
			Header: http.Header{"User-Agent": []string{"TwiceAgent"}},
			Body:   []byte(`{"twice":"data"}`),
		}
		flowAdapter.SetFlow(&px.Flow{Request: req2})

		// previous request data is returned, because SetFlow will not replace the request if it's already set
		assert.Equal(t, "GET", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "http://example.com/twice", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{"User-Agent": []string{"TwiceAgent"}}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(`{"twice":"data"}`), flowAdapter.GetRequest().GetBodyBytes())
	})
}

func TestFlowAdapterSetResponse(t *testing.T) {
	t.Parallel()
	flowAdapter := &FlowAdapter{}
	res := &px.Response{
		StatusCode: 404,
		Header:     http.Header{"Content-Type": []string{"application/xml"}},
		Body:       []byte(`{"error":"not found"}`),
	}
	flowAdapter.SetResponse(res)

	assert.Equal(t, 404, flowAdapter.GetResponse().GetStatusCode())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/xml"}}, flowAdapter.GetResponse().GetHeaders())
	assert.Equal(t, []byte(`{"error":"not found"}`), flowAdapter.GetResponse().GetBodyBytes())
}

func TestFlowAdapterSetFlow(t *testing.T) {
	t.Parallel()
	t.Run("FlowWithRequestAndResponseSet", func(t *testing.T) {
		flow := &px.Flow{
			Request: &px.Request{
				Method: "PUT",
				URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/putflow"},
				Proto:  "HTTP/1.1",
				Header: http.Header{"User-Agent": []string{"PutAgent"}},
				Body:   []byte(`{"putflow":"data"}`),
			},
			Response: &px.Response{
				StatusCode: 201,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       []byte(`{"response":"created"}`),
			},
		}

		flowAdapter := &FlowAdapter{}
		flowAdapter.SetFlow(flow)

		assert.Equal(t, "PUT", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "http://example.com/putflow", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{"User-Agent": []string{"PutAgent"}}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(`{"putflow":"data"}`), flowAdapter.GetRequest().GetBodyBytes())

		assert.Equal(t, 201, flowAdapter.GetResponse().GetStatusCode())
		assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, flowAdapter.GetResponse().GetHeaders())
		assert.Equal(t, []byte(`{"response":"created"}`), flowAdapter.GetResponse().GetBodyBytes())
	})

	t.Run("FlowWithNilRequestAndResponse", func(t *testing.T) {
		t.Parallel()
		flow := &px.Flow{}

		flowAdapter := NewFlowAdapter(flow)

		assert.Equal(t, "", flowAdapter.GetRequest().GetMethod())
		assert.Equal(t, "", flowAdapter.GetRequest().GetURL().String())
		assert.Equal(t, "", flowAdapter.GetRequest().GetProto())
		assert.Equal(t, http.Header{}, flowAdapter.GetRequest().GetHeaders())
		assert.Equal(t, []byte(nil), flowAdapter.GetRequest().GetBodyBytes())

		assert.Equal(t, 0, flowAdapter.GetResponse().GetStatusCode())
		assert.Equal(t, http.Header{}, flowAdapter.GetResponse().GetHeaders())
		assert.Equal(t, []byte(nil), flowAdapter.GetResponse().GetBodyBytes())
	})
}

func TestFlowAdapterSetFlowWithExistingRequestResponse(t *testing.T) {
	t.Parallel()

	initialReq := &px.Request{
		Method: "DELETE",
		URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/deleteflow"},
		Proto:  "HTTP/1.1",
		Header: http.Header{"User-Agent": []string{"InitialAgent"}},
		Body:   []byte(`{"deleteflow":"data"}`),
	}
	initialRes := &px.Response{
		StatusCode: 500,
		Header:     http.Header{"Content-Type": []string{"application/problem+json"}},
		Body:       []byte(`{"error":"server error"}`),
	}

	flowAdapter := &FlowAdapter{}
	flowAdapter.SetRequest(initialReq)
	flowAdapter.SetResponse(initialRes)

	newFlow := &px.Flow{
		Request: &px.Request{
			Method: "PATCH",
			URL:    &url.URL{Scheme: "http", Host: "example.com", Path: "/patchflow"},
			Proto:  "HTTP/1.1",
			Header: http.Header{"User-Agent": []string{"PatchAgent"}},
			Body:   []byte(`{"patchflow":"data"}`),
		},
		Response: &px.Response{
			StatusCode: 202,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"response":"patched"}`),
		},
	}

	flowAdapter.SetFlow(newFlow)

	assert.Equal(t, newFlow, flowAdapter.connectionStats.f)
	assert.Equal(t, "DELETE", flowAdapter.GetRequest().GetMethod())
	assert.Equal(t, "http://example.com/deleteflow", flowAdapter.GetRequest().GetURL().String())
	assert.Equal(t, "HTTP/1.1", flowAdapter.GetRequest().GetProto())
	assert.Equal(t, http.Header{"User-Agent": []string{"InitialAgent"}}, flowAdapter.GetRequest().GetHeaders())
	assert.Equal(t, []byte(`{"deleteflow":"data"}`), flowAdapter.GetRequest().GetBodyBytes())

	assert.Equal(t, 500, flowAdapter.GetResponse().GetStatusCode())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/problem+json"}}, flowAdapter.GetResponse().GetHeaders())
	assert.Equal(t, []byte(`{"error":"server error"}`), flowAdapter.GetResponse().GetBodyBytes())
}
