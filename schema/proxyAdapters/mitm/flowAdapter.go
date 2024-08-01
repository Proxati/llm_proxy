package mitm

import (
	"net/url"

	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters"
	px "github.com/proxati/mitmproxy/proxy"
)

// FlowAdapter implements the proxyAdapters.ProxyFlowReaderAdapter interface
type FlowAdapter struct {
	f   *px.Flow
	req *px.Request
	res *px.Response
}

func NewFlowAdapter(flow *px.Flow) *FlowAdapter {
	if flow == nil {
		return nil
	}
	newAdapter := &FlowAdapter{}
	newAdapter.SetFlow(flow)
	return newAdapter
}

// SetRequest copies the request, to keep the original flow
func (fa *FlowAdapter) SetRequest(req px.Request) {
	fa.req = &req
}

func (fa *FlowAdapter) SetResponse(res px.Response) {
	fa.res = &res
}

func (fa *FlowAdapter) SetFlow(flow *px.Flow) {
	fa.f = flow

	// Only set these if they're not already set
	if fa.req == nil {
		if flow.Request != nil {
			fa.SetRequest(*flow.Request)
		} else {
			// Set an empty request if flow.Request is nil
			fa.SetRequest(px.Request{
				Header: make(map[string][]string),
				URL:    &url.URL{},
			})
		}
	}

	if fa.res == nil {
		if flow.Response != nil {
			fa.SetResponse(*flow.Response)
		} else {
			// Set an empty response if flow.Response is nil
			fa.SetResponse(px.Response{
				Header: make(map[string][]string),
			})
		}
	}
}

func (fa *FlowAdapter) GetRequest() proxyAdapters.RequestReaderAdapter {
	return NewProxyRequestAdapter(fa.req)
}

func (fa *FlowAdapter) GetResponse() proxyAdapters.ResponseReaderAdapter {
	return NewProxyResponseAdapter(fa.res)
}

func (fa *FlowAdapter) GetConnectionStats() proxyAdapters.ConnectionStatsReaderAdapter {
	return NewProxyConnectionStatsAdapter(fa.f)
}
