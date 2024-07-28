package addons

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/providers"
	"github.com/proxati/llm_proxy/v2/schema/proxyAdapters/mitm"
	px "github.com/proxati/mitmproxy/proxy"
)

// APIAuditorAddon log connection and flow
type APIAuditorAddon struct {
	px.BaseAddon
	costCounter *schema.CostCounter
	closed      atomic.Bool
	wg          sync.WaitGroup
	logger      *slog.Logger
}

func (aud *APIAuditorAddon) Response(f *px.Flow) {
	logger := aud.logger.With("URL", f.Request.URL, "StatusCode", f.Response.StatusCode, "ID", f.Id.String())

	if aud.closed.Load() {
		logger.Warn("APIAuditor is being closed, not processing request")
		return
	}
	if f.Response == nil {
		logger.Debug("skipping accounting for nil response")
		return
	}

	aud.wg.Add(1) // for blocking this addon during shutdown in .Close()
	go func() {
		defer aud.wg.Done()
		<-f.Done()

		// only account when the request domain is supported
		reqHostname := f.Request.URL.Hostname()
		_, shouldAudit := providers.API_Hostnames[reqHostname]
		if !shouldAudit {
			logger.Debug("skipping accounting for unsupported API")
			return
		}

		// Only account when receiving good response codes
		_, shouldAccount := cacheOnlyResponseCodes[f.Response.StatusCode]
		if !shouldAccount {
			logger.Debug("skipping accounting for non-200 response")
			return
		}

		// convert the request to an internal TrafficObject
		reqAdapter := mitm.NewProxyRequestAdapter(f.Request) // generic wrapper for the mitm request

		tObjReq, err := schema.NewProxyRequest(reqAdapter, []string{})
		if err != nil {
			logger.Error("error creating TrafficObject from request", "error", err)
			return
		}

		// convert the response to an internal TrafficObject
		respAdapter := mitm.NewProxyResponseAdapter(f.Response) // generic wrapper for the mitm response

		tObjResp, err := schema.NewProxyResponse(respAdapter, []string{})
		if err != nil {
			logger.Error("error creating TrafficObject from response", "error", err)
			return
		}

		// account the cost, TODO: returns what?
		auditOutput, err := aud.costCounter.Add(*tObjReq, *tObjResp)
		if err != nil {
			logger.Error("error accounting response", "error", err)
		}

		// TODO Improve this output format:
		fmt.Println(auditOutput)
	}()
}

func (aud *APIAuditorAddon) Close() error {
	if !aud.closed.Swap(true) {
		aud.logger.Debug("Waiting for APIAuditor shutdown...")
		aud.wg.Wait()
	}

	return nil
}

func NewAPIAuditor(logger *slog.Logger) *APIAuditorAddon {
	aud := &APIAuditorAddon{
		costCounter: schema.NewCostCounterDefaults(),
		logger:      logger.WithGroup("addons.APIAuditorAddon"),
	}
	aud.closed.Store(false) // initialize as open
	return aud
}
