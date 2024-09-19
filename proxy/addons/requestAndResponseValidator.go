package addons

import (
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/proxati/llm_proxy/v2/proxy/addons/helpers"
	px "github.com/proxati/mitmproxy/proxy"
)

type RequestAndResponseValidator struct {
	px.BaseAddon
	logger *slog.Logger
	wg     sync.WaitGroup
	closed atomic.Bool
}

func NewRequestAndResponseValidator(logger *slog.Logger) *RequestAndResponseValidator {
	return &RequestAndResponseValidator{logger: logger.WithGroup("addons.RequestAndResponseValidator")}
}

// Request validates the request, and does not use the normal logger, because
// the request may not have all the necessary fields needed to log.
func (c *RequestAndResponseValidator) Request(f *px.Flow) {
	if c.closed.Load() {
		helpers.GenerateClosedResponse(c.logger, f)
		return
	} else {
		c.wg.Add(1)
		defer c.wg.Done()
	}

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

// Response validates the response, and does not use the normal logger, because
// the response may not have all the necessary fields needed to log.
func (c *RequestAndResponseValidator) Response(f *px.Flow) {
	if !c.closed.Load() {
		// if the addon is NOT closed, then add to the wait group
		c.wg.Add(1)
		defer c.wg.Done()
	}

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

func (d *RequestAndResponseValidator) Close() error {
	if !d.closed.Swap(true) {
		d.logger.Debug("Closing...")
		d.wg.Wait()
	}

	return nil
}

func (d *RequestAndResponseValidator) String() string {
	return "RequestAndResponseValidator"
}
