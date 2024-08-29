package mitm

import (
	"net/url"

	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters"
	px "github.com/proxati/mitmproxy/proxy"
)

// FlowAdapter implements the proxyAdapters.ProxyFlowReaderAdapter interface
type FlowAdapter struct {
	connectionStats *ConnectionStatsAdapter
	req             *ProxyRequestAdapter
	res             *ProxyResponseAdapter
}

// NewFlowAdapter creates a new flow adapter object
func NewFlowAdapter(flow *px.Flow) *FlowAdapter {
	if flow == nil {
		return nil
	}
	if flow.Request == nil {
		flow.Request = &px.Request{
			Header: map[string][]string{},
			URL:    &url.URL{},
		}
	}
	if flow.Response == nil {
		flow.Response = &px.Response{
			Header: map[string][]string{},
		}
	}

	newAdapter := &FlowAdapter{}
	newAdapter.SetFlow(flow)
	return newAdapter
}

// SetRequest copies the request, to keep the original flow
func (fa *FlowAdapter) SetRequest(req *px.Request) {
	fa.req = NewProxyRequestAdapter(req)
}

// SetResponse copies the response, to keep the original flow
func (fa *FlowAdapter) SetResponse(res *px.Response) {
	fa.res = NewProxyResponseAdapter(res)
}

// SetFlow sets the connectionStats, request, and response (but only if the request and response
// weren't previously set). If SetRequest or SetResponse was run before this, the request/response
// in this flow will be ignored.
func (fa *FlowAdapter) SetFlow(flow *px.Flow) {
	fa.connectionStats = NewProxyConnectionStatsAdapter(flow)

	// Only set these if they're not already set
	if fa.req == nil {
		fa.SetRequest(flow.Request)
	}
	if fa.res == nil {
		fa.SetResponse(flow.Response)
	}
}

// GetRequest returns the request adapter, to implement the FlowReaderAdapter interface
func (fa *FlowAdapter) GetRequest() proxyAdapters.RequestReaderAdapter {
	return fa.req
}

// GetResponse returns the response adapter, to implement the FlowReaderAdapter interface
func (fa *FlowAdapter) GetResponse() proxyAdapters.ResponseReaderAdapter {
	return fa.res
}

// GetConnectionStats returns the connection stats adapter, to implement the FlowReaderAdapter interface
func (fa *FlowAdapter) GetConnectionStats() proxyAdapters.ConnectionStatsReaderAdapter {
	return fa.connectionStats
}
