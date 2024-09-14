package helpers

import (
	"log/slog"
	"net/http"

	"github.com/proxati/llm_proxy/v2/schema/headers"
	px "github.com/proxati/mitmproxy/proxy"
)

// RequestClosed is the function used by the Request method when the addon is closed. It doesn't
// return anything, but instead attaches a 503 response to the flow, and sets a few headers on
// the response. When the proxy sees the response != nil, it will skip the rest of the addons.
func RequestClosed(logger *slog.Logger, f *px.Flow) {
	logger.WithGroup("closed").Warn("sending a 503 response to client, because this addon is being closed")
	f.Response = &px.Response{
		StatusCode: http.StatusServiceUnavailable,
		Body:       []byte("LLM_Proxy is not available"),
		Header: http.Header{
			"Content-Type":            {"text/plain"},
			headers.CacheStatusHeader: {headers.CacheStatusValueSkip}, // Not ideal to always add this here
			"Connection":              {"close"},
		},
	}
}
