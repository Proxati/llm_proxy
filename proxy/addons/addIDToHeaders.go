package addons

import (
	"github.com/proxati/llm_proxy/v2/version"
	px "github.com/proxati/mitmproxy/proxy"
)

const (
	idHeader  = "X-Llm_proxy-id"
	idVersion = "X-Llm_proxy-version"
)

type AddIDToHeaders struct {
	px.BaseAddon
}

func NewAddIDToHeaders() *AddIDToHeaders {
	return &AddIDToHeaders{}
}

func (c *AddIDToHeaders) Response(f *px.Flow) {
	f.Response.Header.Add(idHeader, f.Id.String())
	f.Response.Header.Add(idVersion, version.String())
}
