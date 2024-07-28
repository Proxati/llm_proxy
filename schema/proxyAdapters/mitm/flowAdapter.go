package mitm

import (
	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters"
	px "github.com/proxati/mitmproxy/proxy"
)

// FlowAdapter implements the proxyAdapters.ProxyFlowReaderAdapter interface
type FlowAdapter struct {
	f *px.Flow
}

func NewFlowAdapter(f *px.Flow) *FlowAdapter {
	if f == nil {
		return nil
	}

	return &FlowAdapter{f: f}
}

func (fa *FlowAdapter) GetRequest() proxyAdapters.RequestReaderAdapter {
	return NewProxyRequestAdapter(fa.f.Request)
}

func (fa *FlowAdapter) GetResponse() proxyAdapters.ResponseReaderAdapter {
	return NewProxyResponseAdapter(fa.f.Response)
}

func (fa *FlowAdapter) GetConnectionStats() proxyAdapters.ConnectionStatsReaderAdapter {
	return NewProxyConnectionStatsAdapter(fa.f)
}
