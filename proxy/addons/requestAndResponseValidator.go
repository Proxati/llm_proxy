package addons

import (
	"log/slog"
	"net/http"

	px "github.com/proxati/mitmproxy/proxy"
)

func getProxyIDSafe(f *px.Flow) string {
	if f.ConnContext != nil && f.ConnContext.ClientConn != nil {
		return f.ConnContext.ID().String()
	}
	return ""
}

type RequestAndResponseValidator struct {
	px.BaseAddon
	logger *slog.Logger
}

func NewRequestAndResponseValidator(logger *slog.Logger) *RequestAndResponseValidator {
	return &RequestAndResponseValidator{logger: logger.WithGroup("addons.RequestAndResponseValidator")}
}

func (c *RequestAndResponseValidator) Request(f *px.Flow) {
	if f.Request != nil {
		if f.Request.URL == nil || f.Request.URL.String() == "" {
			c.logger.Error(
				"request URL is nil or empty",
				"client.ID", f.Id.String(),
				"proxy.ID", getProxyIDSafe(f),
			)
			f.Response = &px.Response{
				StatusCode: http.StatusBadRequest,
				Body:       []byte("Request URL is empty"),
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
			}
			return
		}
		return
	}

	c.logger.Error(
		"request is nil",
		"client.ID", f.Id.String(),
		"proxy.ID", getProxyIDSafe(f),
	)
	f.Response = &px.Response{
		StatusCode: http.StatusBadRequest,
		Body:       []byte("Request is nil"),
		Header: http.Header{
			"Content-Type": []string{"text/plain"},
		},
	}
}

func (c *RequestAndResponseValidator) Response(f *px.Flow) {
	if f.Response != nil {
		return
	}
	c.logger.Error(
		"response is nil",
		"URL", f.Request.URL,
		"client.ID", f.Id.String(),
		"proxy.ID", getProxyIDSafe(f),
	)
	f.Response = &px.Response{
		StatusCode: http.StatusBadGateway,
		Body:       []byte("Response is nil"),
		Header: http.Header{
			"Content-Type": []string{"text/plain"},
		},
	}
}
