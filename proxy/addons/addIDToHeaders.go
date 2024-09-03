package addons

import (
	"github.com/proxati/llm_proxy/v2/schema/headers"
	"github.com/proxati/llm_proxy/v2/version"
	px "github.com/proxati/mitmproxy/proxy"
)

type AddIDToHeaders struct {
	px.BaseAddon
}

func NewAddIDToHeaders() *AddIDToHeaders {
	return &AddIDToHeaders{}
}

func (c *AddIDToHeaders) Response(f *px.Flow) {
	f.Response.Header.Add(headers.ProxyID, f.Id.String())
	f.Response.Header.Add(headers.Version, version.String())
}
