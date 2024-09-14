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
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters/mitm"
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
}

func (a *TrafficTransformerAddon) Request(f *px.Flow) {
	a.wg.Add(1)
	defer a.wg.Done()
	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Request")

	if a.closed.Load() {
		logger.ErrorContext(a.ctx, "Addon is closed, not processing request")
		helpers.RequestClosed(a.logger, f)
		return
	}

	logger.DebugContext(a.ctx, "Starting transformations")

	req := mitm.NewProxyRequestAdapter(f.Request)
	if req == nil {
		logger.ErrorContext(a.ctx, "Failed to create request adapter")
		helpers.ProxyError(a.logger, f)
		return
	}

	var newReq proxyadapters.RequestReaderAdapter
	var newResp proxyadapters.ResponseReaderAdapter
	logger.DebugContext(a.ctx, "debug...", "req", req, "newReq", newReq, "newResp", newResp)

	for name, providers := range a.requestProviders {
		logger := logger.With("name", name)
		for _, provider := range providers {
			logger.DebugContext(a.ctx, "Communicating with transformer")
			Tcfg := provider.GetTransformerConfig()

			// TODO: handle multiple providers with the same service name
			// - failover / backup ?
			// - random / round robin ?
			req, resp, err := provider.Transform(a.ctx, req, newReq, newResp)
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

	if newReq != nil {
		// TODO: update the flow with the new request
		logger.DebugContext(a.ctx, "request object updated from transformers")
	}

	if newResp != nil {
		// TODO: update the flow with the new response
		logger.DebugContext(a.ctx, "response object updated from transformers")
	}

	logger.DebugContext(a.ctx, "transformations completed for request")
}

func (a *TrafficTransformerAddon) Response(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Response")
	a.wg.Add(1)
	defer a.wg.Done()

	if f.Response != nil && (f.Response.StatusCode < 100 || f.Response.StatusCode > 999) {
		// defense to prevent a misbehaving transformer from setting an invalid status code
		logger.ErrorContext(a.ctx, "Invalid StatusCode in response", "StatusCode", f.Response.StatusCode)
		f.Response.StatusCode = http.StatusInternalServerError
	}

	// TODO: response body editing here
	logger.DebugContext(a.ctx, "Done editing response body")
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
			logger.WithGroup(name)
			for _, provider := range providers {
				defer a.wg.Done()

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

func loadTransformerProviders(logger *slog.Logger, ctx context.Context, wg *sync.WaitGroup, transformersConfig []*config.Transformer) (map[string][]transformers.Provider, error) {
	providers := make(map[string][]transformers.Provider)
	errs := []error{}
	for _, t := range transformersConfig {
		var prov transformers.Provider
		var err error

		switch strings.ToLower(t.URL.Scheme) {
		/*
			case "file":
				prov, err = transformers.NewFileProvider(t.URL)
			case "grpc":
				prov, err = transformers.NewGrpcProvider(t.URL)
		*/
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
			wg.Add(1)
			providers[t.Name] = append(providers[t.Name], prov)
		}
	}

	for k, v := range providers {
		if len(v) > 1 {
			logger.Warn("TODO: Multiple providers configured with the same service name, feature is currently unimplemented", "name", k)
		}
	}

	return providers, nil
}

func NewTrafficTransformerAddon(
	logger *slog.Logger,
	requestTransformer []*config.Transformer,
	responseTransformer []*config.Transformer,
) (*TrafficTransformerAddon, error) {
	logger = logger.WithGroup("addons.TrafficTransformerAddon")
	wg := &sync.WaitGroup{}

	// create a new context for the addon, this handles shutdown of the providers and the addon itself
	ctx, cancel := context.WithCancel(context.Background())

	// load the providers for the various configured transformers
	reqProv, reqErr := loadTransformerProviders(logger, ctx, wg, requestTransformer)
	respProv, respErr := loadTransformerProviders(logger, ctx, wg, responseTransformer)
	if reqErr != nil || respErr != nil {
		cancel()
		return nil, errors.Join(errors.New("unable to load transformer(s)"), reqErr, respErr)
	}
	logger.Debug("Loaded request transformers", "count", len(reqProv))
	logger.Debug("Loaded response transformers", "count", len(respProv))

	ta := &TrafficTransformerAddon{
		logger:            logger,
		requestProviders:  reqProv,
		responseProviders: respProv,
		ctx:               ctx,
		cancel:            cancel,
	}

	return ta, nil
}
