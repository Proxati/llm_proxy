package addons

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/proxati/llm_proxy/v2/config"
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
	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Request")
	a.wg.Add(1)
	defer a.wg.Done()

	fa := mitm.NewFlowAdapter(f)
	ctx := context.WithValue(a.ctx, "flow", fa)
	var newReq proxyadapters.ResponseReaderAdapter
	var newResp proxyadapters.RequestReaderAdapter

	for name, providers := range a.requestProviders {
		logger := logger.With("name", name)
		for _, provider := range providers {
			logger.Debug("Communicating with transformer")
			// TODO: handle multiple providers with the same service name
			// - failover / backup ?
			// - random / round robin ?
			req, resp, err := provider.Transform(ctx, fa, newReq, newResp)
			if err != nil {
				logger.Error("Failed to transform request", "error", err)
				continue
			}
			if req != nil {
				newReq = req
				logger.Debug("Transformer updated request", "request", req)
			}
			if resp != nil {
				newResp = resp
				logger.Debug("Transformer updated response", "response", resp)
			}
			logger.Debug("Done communicating with transformer")
		}
	}

	if newReq != nil {
		// TODO: update the flow with the new request
		logger.Debug("Done editing request")
	}

	if newResp != nil {
		// TODO: update the flow with the new response
		logger.Debug("Done editing response")
	}

	logger.Debug("Done with transformations")
}

func (a *TrafficTransformerAddon) Response(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(a.logger, f).WithGroup("Response")
	a.wg.Add(1)
	defer a.wg.Done()

	// TODO: response body editing here
	logger.Debug("Done editing response body")
}

func (a *TrafficTransformerAddon) String() string {
	return fmt.Sprintf(
		"TrafficTransformerAddon (Request Providers: %d, Response Providers: %d)",
		len(a.requestProviders), len(a.responseProviders))
}

func (a *TrafficTransformerAddon) Close() error {
	if !a.closed.Swap(true) {
		a.logger.Debug("Closing...")
		a.cancel()
		a.wg.Wait()
	}

	return nil
}

func loadTransformers(logger *slog.Logger, transformersConfig []*config.Transformer) (map[string][]transformers.Provider, error) {
	providers := make(map[string][]transformers.Provider)
	errs := []error{}
	for _, t := range transformersConfig {
		var prov transformers.Provider
		var err error

		switch t.URL.Scheme {
		/*
			case "file":
				prov, err = transformers.NewFileProvider(t.URL)
			case "grpc":
				prov, err = transformers.NewGrpcProvider(t.URL)
		*/
		case "http", "https":
			prov, err = transformers.NewHttpProvider(logger, t)
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

	return providers, nil
}

func NewTrafficTransformerAddon(
	logger *slog.Logger,
	requestTransformer []*config.Transformer,
	responseTransformer []*config.Transformer,
) (*TrafficTransformerAddon, error) {
	logger = logger.WithGroup("addons.TrafficTransformerAddon")

	// load the providers for the various configured transformers
	reqProv, reqErr := loadTransformers(logger, requestTransformer)
	respProv, respErr := loadTransformers(logger, responseTransformer)
	if reqErr != nil || respErr != nil {
		return nil, errors.Join(errors.New("unable to load transformer(s)"), reqErr, respErr)
	}
	logger.Debug("Loaded request transformers", "count", len(reqProv))
	logger.Debug("Loaded response transformers", "count", len(respProv))

	ctx, cancel := context.WithCancel(context.Background())

	ta := &TrafficTransformerAddon{
		logger:            logger,
		requestProviders:  reqProv,
		responseProviders: respProv,
		ctx:               ctx,
		cancel:            cancel,
	}

	return ta, nil
}
