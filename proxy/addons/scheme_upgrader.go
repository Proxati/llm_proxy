package addons

import (
	"log/slog"

	px "github.com/proxati/mitmproxy/proxy"
)

const (
	upgradedHeader = "X-Llm_proxy-scheme-upgraded"
)

type SchemeUpgrader struct {
	px.BaseAddon
}

func (c *SchemeUpgrader) Request(f *px.Flow) {
	logger := slog.With("addon", "SchemeUpgrader.Request", "URL", f.Request.URL, "ID", f.Id.String())

	// upgrade to https
	if f.Request.URL.Scheme == "https" {
		logger.Debug("Upgrading URL scheme from http to https not needed for URL")
		return
	}

	// add a header to the request to indicate that the scheme was upgraded
	f.Request.Header.Add(upgradedHeader, "true")

	// upgrade the connection from http to https, so when sent upstream it will be encrypted
	f.Request.URL.Scheme = "https"
}

func (c *SchemeUpgrader) Response(f *px.Flow) {
	if f.Request.Header.Get(upgradedHeader) != "" {
		f.Response.Header.Add(upgradedHeader, "true")
	}
}
