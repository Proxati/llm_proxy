package helpers

import (
	"log/slog"
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"
)

func ProxyError(logger *slog.Logger, f *px.Flow) {
	logger.Error("internal proxy error")
	f.Response = &px.Response{
		StatusCode: http.StatusBadGateway,
		Body:       []byte("LLM_Proxy internal error"),
		Header: http.Header{
			"Content-Type": {"text/plain"},
			"Connection":   {"close"},
		},
	}
}
