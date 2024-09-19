package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons"
	"github.com/proxati/llm_proxy/v2/proxy/addons/helpers"
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
	logger         *slog.Logger
	closed         atomic.Bool
}

// NewMetaAddon creates a new MetaAddon with the given config and addons. The order of the addons
// is important, as they will be processed in the order they are given. The first addon to return
// with a response will be the final addon to process the request. Logging will be handled in a
// separate system, and not in the addons (so the response can be captured).
func newMetaAddon(logger *slog.Logger, cfg *config.Config, addons ...px.Addon) *metaAddon {
	m := &metaAddon{
		cfg:    cfg,
		logger: logger.WithGroup("MetaAddon"),
	}

	// iterate so the addons can be type asserted and added to the correct field
	for _, a := range addons {
		if err := m.addAddon(a); err != nil {
			m.logger.Error("could not add the metaAddon", "error", err)
		}
	}

	return m
}

func (ma *metaAddon) Close() error {
	if !ma.closed.Swap(true) {
		ma.logger.Debug("Closing all sub-addons...")
		for _, a := range ma.closableAddons {
			logger := ma.logger.With("addonName", a.String())
			if err := a.Close(); err != nil {
				logger.Error("error while closing", "error", err)
				continue
			}
			logger.Debug("Closed addon")
		}
	}

	return nil
}

func (*metaAddon) String() string {
	return "metaAddon"
}

func (ma *metaAddon) addAddon(a any) error {
	if a == nil {
		ma.logger.Debug("Skipping add for nil addon")
		return nil
	}

	myAddon, ok := a.(addons.ClosableAddon)
	if ok {
		ma.closableAddons = append(ma.closableAddons, myAddon) // for closing later
		ma.logger.Debug("Loaded closable addon", "addonName", myAddon.String())
	}

	mitmAddon, ok := a.(px.Addon)
	if ok {
		// the addon is a valid mitmproxy addon, but it lacks a .String() method so we can't log it
		ma.mitmAddons = append(ma.mitmAddons, mitmAddon)
		return nil
	}

	return fmt.Errorf("invalid addon type: %T", a)
}

func (addon *metaAddon) ClientConnected(client *px.ClientConn) {
	if addon.closed.Load() {
		addon.logger.Warn("skipping addons for ClientConnected, metaAddon is being closed")
		return
	}

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
	if addon.closed.Load() {
		addon.logger.Warn("skipping addons for Requestheaders, metaAddon is being closed")
		helpers.GenerateClosedResponse(addon.logger, flow)
		return
	}

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
