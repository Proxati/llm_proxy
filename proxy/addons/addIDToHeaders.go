package addons

import (
	px "github.com/proxati/mitmproxy/proxy"
)

const (
	idHeader = "X-Llm_proxy-id"
)

type AddIDToHeaders struct {
	px.BaseAddon
}

func NewAddIDToHeaders() *AddIDToHeaders {
	return &AddIDToHeaders{}
}

func (c *AddIDToHeaders) Response(f *px.Flow) {
	f.Response.Header.Add(idHeader, f.Id.String())
}
