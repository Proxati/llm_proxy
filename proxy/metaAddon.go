package proxy

import (
	"fmt"
	"io"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons"
	px "github.com/proxati/mitmproxy/proxy"
)

// metaAddon is a meta addon that is the only addon loaded by the upstream library, and all
// of our internal addons are processed here. This is the first step of an abstraction layer,
// so the upstream library can be replaced with a different one in the future.
type metaAddon struct {
	px.BaseAddon
	cfg            *config.Config
	mitmAddons     []px.Addon
	closableAddons []addons.ClosableAddon
}

// NewMetaAddon creates a new MetaAddon with the given config and addons. The order of the addons
// is important, as they will be processed in the order they are given. The first addon to return
// with a response will be the final addon to process the request. Logging will be handled in a
// separate system, and not in the addons (so the response can be captured).
func newMetaAddon(cfg *config.Config, addons ...px.Addon) *metaAddon {
	m := &metaAddon{cfg: cfg}

	// iterate so the addons can be type asserted and added to the correct field
	for _, a := range addons {
		if err := m.addAddon(a); err != nil {
			sLogger.Error("could not add the metaAddon", "error", err)
		}
	}

	return m
}

func (addon *metaAddon) Close() error {
	for _, a := range addon.closableAddons {
		if err := a.Close(); err != nil {
			sLogger.Error("could not close the addon", "error", err)
		}
		sLogger.Debug("Closed addon", "addonName", a.String())
	}

	return nil
}

func (addon *metaAddon) String() string {
	return "metaAddon"
}

func (addon *metaAddon) addAddon(a any) error {
	if a == nil {
		sLogger.Debug("Skipping add for nil addon")
		return nil
	}

	myAddon, ok := a.(addons.ClosableAddon)
	if ok {
		addon.closableAddons = append(addon.closableAddons, myAddon) // for closing later
		sLogger.Debug("Loaded closable addon", "addonName", myAddon.String())
	}

	mitmAddon, ok := a.(px.Addon)
	if ok {
		// the addon is a valid mitmproxy addon, but it lacks a .String() method so we can't log it
		addon.mitmAddons = append(addon.mitmAddons, mitmAddon)
		return nil
	}

	return fmt.Errorf("invalid addon type: %T", a)
}

func (addon *metaAddon) ClientConnected(client *px.ClientConn) {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for client mutations, so just iterate peacefully
		a.ClientConnected(client)
	}

	// TODO: add a logger here
}

func (addon *metaAddon) ClientDisconnected(client *px.ClientConn) {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for client mutations, so just iterate peacefully
		a.ClientDisconnected(client)
	}

	// TODO: add a logger here
}

func (addon *metaAddon) ServerConnected(ctx *px.ConnContext) {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for context mutations, so just iterate peacefully
		a.ServerConnected(ctx)
	}

	// TODO: add a logger here
}

func (addon *metaAddon) ServerDisconnected(ctx *px.ConnContext) {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for context mutations, so just iterate peacefully
		a.ServerDisconnected(ctx)
	}

	// TODO: add a logger here
}

func (addon *metaAddon) TlsEstablishedServer(ctx *px.ConnContext) {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for context mutations, so just iterate peacefully
		a.TlsEstablishedServer(ctx)
	}

	// TODO: add a logger here
}

func (addon *metaAddon) Requestheaders(flow *px.Flow) {
	for _, a := range addon.mitmAddons {
		a.Requestheaders(flow)
		if flow.Response != nil {
			// the response has been set, stop processing addons
			break
		}
	}

	// TODO: add a logger here
}

func (addon *metaAddon) Request(flow *px.Flow) {
	for _, a := range addon.mitmAddons {
		a.Request(flow)
		if flow.Response != nil {
			// the response has been set, stop processing addons
			break
		}
	}

	// TODO: add a logger here
}

func (addon *metaAddon) Responseheaders(flow *px.Flow) {
	for _, a := range addon.mitmAddons {
		a.Responseheaders(flow)

		if flow.Response != nil && flow.Response.Body != nil {
			// the response body has been set, stop processing addons
			break
		}
	}

	// TODO: add a logger here
}

func (addon *metaAddon) Response(flow *px.Flow) {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for flow mutations, so just iterate peacefully
		a.Response(flow)
	}

	// TODO: add a logger here
}

func (addon *metaAddon) StreamRequestModifier(flow *px.Flow, in io.Reader) io.Reader {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for flow mutations, so just iterate peacefully
		in = a.StreamRequestModifier(flow, in)
	}

	return in
}

func (addon *metaAddon) StreamResponseModifier(flow *px.Flow, in io.Reader) io.Reader {
	for _, a := range addon.mitmAddons {
		// the caller for this method doesn't check for flow mutations, so just iterate peacefully
		in = a.StreamResponseModifier(flow, in)
	}

	return in
}
