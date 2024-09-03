package addons

import (
	"log/slog"

	"github.com/proxati/llm_proxy/v2/schema/headers"
	px "github.com/proxati/mitmproxy/proxy"
)

type SchemeUpgrader struct {
	px.BaseAddon
	logger *slog.Logger
}

func NewSchemeUpgrader(logger *slog.Logger) *SchemeUpgrader {
	return &SchemeUpgrader{
		logger: logger.WithGroup("addons.SchemeUpgrader"),
	}
}

func (c *SchemeUpgrader) Request(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(c.logger, f).WithGroup("Request")

	// upgrade to https
	if f.Request.URL.Scheme == "https" {
		logger.Debug("Upgrading URL scheme from http to https not needed for URL")
		return
	}

	// add a header to the request to indicate that the scheme was upgraded
	f.Request.Header.Add(headers.SchemeUpgraded, "true")

	// upgrade the connection from http to https, so when sent upstream it will be encrypted
	f.Request.URL.Scheme = "https"
}

func (c *SchemeUpgrader) Response(f *px.Flow) {
	if f.Request.Header.Get(headers.SchemeUpgraded) != "" {
		f.Response.Header.Add(headers.SchemeUpgraded, "true")
	}
}
