package addons

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons/helpers"
	"github.com/proxati/llm_proxy/v2/proxy/addons/transformers"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters/mitm"
	"github.com/proxati/llm_proxy/v2/schema/utils"
	px "github.com/proxati/mitmproxy/proxy"
)

type TrafficTransformerAddon struct {
	px.BaseAddon
	wg                sync.WaitGroup
	closed            atomic.Bool
	logger            *slog.Logger
	requestProviders  map[string][]transformers.Provider
	responseProviders map[string][]transformers.Provider
	ctx               context.Context
	cancel            context.CancelFunc
	headerFilter      *config.HeaderFilterGroup
}

func NewTrafficTransformerAddon(
	logger *slog.Logger,
	requestTransformer []*config.Transformer,
	responseTransformer []*config.Transformer,
) (*TrafficTransformerAddon, error) {
	logger = logger.WithGroup("addons.TrafficTransformerAddon")

	// create a new context for the addon, this handles shutdown of the providers and the addon itself
	ctx, cancel := context.WithCancel(context.Background())

	// load the providers for the various configured transformers
	reqProv, reqErr := loadTransformerProviders(logger, ctx, requestTransformer)
	respProv, respErr := loadTransformerProviders(logger, ctx, responseTransformer)
	if reqErr != nil || respErr != nil {
		cancel()
		return nil, errors.Join(errors.New("unable to load transformer(s)"), reqErr, respErr)
	}
	logger.Debug("Loaded request transformers", "count", len(reqProv))
	logger.Debug("Loaded response transformers", "count", len(respProv))

	ta := &TrafficTransformerAddon{
		wg:                sync.WaitGroup{},
		closed:            atomic.Bool{},
		logger:            logger,
		requestProviders:  reqProv,
		responseProviders: respProv,
		ctx:               ctx,
		cancel:            cancel,
		headerFilter:      config.NewHeaderFilterGroup("TODO", []string{}, []string{}),
	}

	return ta, nil
}

func (a *TrafficTransformerAddon) Request(f *px.Flow) {
	a.wg.Add(1)
	defer a.wg.Done()

	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Request")
	defer logger.DebugContext(a.ctx, "transformations completed")

	if a.closed.Load() {
		logger.ErrorContext(a.ctx, "Addon is closed, not processing request")
		helpers.RequestClosed(a.logger, f)
		return
	}

	reqAdapter := mitm.NewProxyRequestAdapter(f.Request)
	if reqAdapter == nil {
		logger.ErrorContext(a.ctx, "Failed to create request adapter")
		helpers.ProxyError(a.logger, f)
		return
	}

	oldReq, err := schema.NewProxyRequest(reqAdapter, a.headerFilter)
	if err != nil {
		logger.ErrorContext(a.ctx, "Failed to create request object", "error", err)
		helpers.ProxyError(a.logger, f)
		return
	}

	logger.DebugContext(a.ctx, "Starting transformations")
	var oldResp *schema.ProxyResponse // there should never be an old response object in the request phase
	var newReq *schema.ProxyRequest
	var newResp *schema.ProxyResponse
	for name, providers := range a.requestProviders {
		logger := logger.With("name", name)
		for _, provider := range providers {
			logger.DebugContext(a.ctx, "Communicating with transformer")
			Tcfg := provider.GetTransformerConfig()

			// TODO: handle multiple providers with the same service name
			// - failover / backup ?
			// - random / round robin ?
			ctx, cancel := context.WithTimeout(a.ctx, Tcfg.ResponseTimeout)
			req, resp, err := provider.Transform(ctx, logger, oldReq, oldResp, newReq, newResp)
			cancel()

			if err != nil {
				logger.ErrorContext(a.ctx, "Failed to transform request", "error", err)
				if Tcfg.FailureMode == config.FailureModeHard {
					helpers.ProxyError(a.logger, f)
					return
				}
				continue
			}

			if req != nil {
				newReq = req
				logger.DebugContext(a.ctx, "Transformer updated request", "request", req)
			}
			if resp != nil {
				newResp = resp
				logger.DebugContext(a.ctx, "Transformer updated response", "response", resp)
			}

			logger.DebugContext(a.ctx, "Done communicating with transformer")
		}
	}
	logger.DebugContext(a.ctx, "Transformer execution completed, checking and merging results...",
		"reqObj", oldReq, "newReq", newReq, "newResp", newResp)

	if newReq != nil {
		contentEncoding := f.Request.Header.Get("Content-Encoding")
		oldReq.Merge(newReq)

		encodedBody, encodingType, err := utils.EncodeBody(oldReq.GetBodyBytes(), contentEncoding)
		if err != nil {
			logger.ErrorContext(a.ctx, "Failed to encode request body", "error", err)
			helpers.ProxyError(a.logger, f)
			return
		}

		f.Request.Method = oldReq.Method
		f.Request.URL = oldReq.URL
		f.Request.Proto = oldReq.Proto
		f.Request.Header = oldReq.Header
		f.Request.Body = encodedBody
		if encodingType != "" {
			f.Request.Header.Set("Content-Encoding", encodingType)
		}
		f.Request.Header.Set("Content-Length", fmt.Sprintf("%d", len(encodedBody)))

		logger.DebugContext(a.ctx, "Request object updated")
	}

	if newResp != nil {
		// a new response object was created, update the flow with the response
		newMitmResp, err := mitm.ToProxyResponse(newResp, f.Response.Header.Get("Accept-Encoding"))
		if err != nil {
			logger.ErrorContext(a.ctx, "Failed to prepare response object", "error", err)
			helpers.ProxyError(a.logger, f)
			return
		}
		f.Response = newMitmResp
		logger.DebugContext(a.ctx, "response object added", "response", newMitmResp)
		return
	}
}

func (a *TrafficTransformerAddon) Response(f *px.Flow) {
	a.wg.Add(1)
	defer a.wg.Done()

	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Response")
	defer logger.DebugContext(a.ctx, "transformations completed")

	if f.Response != nil && (f.Response.StatusCode < 100 || f.Response.StatusCode > 999) {
		// defense to prevent a misbehaving transformer from setting an invalid status code
		logger.ErrorContext(a.ctx, "Invalid StatusCode in response", "StatusCode", f.Response.StatusCode)
		f.Response.StatusCode = http.StatusInternalServerError
	}

	// TODO: response body editing here
}

func (a *TrafficTransformerAddon) String() string {
	return fmt.Sprintf(
		"TrafficTransformerAddon (Request Providers: %d, Response Providers: %d)",
		len(a.requestProviders), len(a.responseProviders))
}

func (a *TrafficTransformerAddon) closeProviders() error {
	logger := a.logger.WithGroup("closeProviders")
	errs := []error{}

	providerMaps := []map[string][]transformers.Provider{
		a.requestProviders,
		a.responseProviders,
	}

	for _, providersMap := range providerMaps {
		for name, providers := range providersMap {
			logger = logger.WithGroup(name)
			for _, provider := range providers {
				logger = logger.With("provider", provider)
				logger.DebugContext(a.ctx, "Starting close")
				if err := provider.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close provider: %w", err))
				}
				logger.DebugContext(a.ctx, "Closed provider")
			}
		}
	}

	return errors.Join(errs...)
}

func (a *TrafficTransformerAddon) Close() error {
	if !a.closed.Swap(true) {
		a.logger.DebugContext(a.ctx, "Closing...")
		a.cancel()
		err := a.closeProviders()
		if err != nil {
			a.logger.ErrorContext(a.ctx, "Failed to close one or more providers", "error", err)
		}

		a.wg.Wait()
	}

	return nil
}

func loadTransformerProviders(logger *slog.Logger, ctx context.Context, transformersConfig []*config.Transformer) (map[string][]transformers.Provider, error) {
	providers := make(map[string][]transformers.Provider)
	errs := []error{}
	for _, t := range transformersConfig {
		var prov transformers.Provider
		var err error

		switch strings.ToLower(t.URL.Scheme) {
		case "file":
			prov, err = transformers.NewFileProvider(logger, ctx, t)
		case "http", "https":
			prov, err = transformers.NewHttpProvider(logger, ctx, t)
		default:
			err = fmt.Errorf("unsupported scheme: %s", t.URL.Scheme)
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create provider %s: %w", t.Name, err))
			continue
		}

		if prov != nil {
			providers[t.Name] = append(providers[t.Name], prov)
		}
	}

	for k, v := range providers {
		if len(v) > 1 {
			logger.Warn("TODO: Multiple providers configured with the same service name, feature is currently unimplemented", "name", k)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return providers, nil
}
