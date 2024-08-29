package addons

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/proxati/llm_proxy/v2/schema/providers"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters/mitm"
	px "github.com/proxati/mitmproxy/proxy"
)

// APIAuditorAddon log connection and flow
type APIAuditorAddon struct {
	px.BaseAddon
	costCounter *schema.CostCounter
	closed      atomic.Bool
	wg          sync.WaitGroup
	logger      *slog.Logger
	auditLogger *slog.Logger
}

func (aud *APIAuditorAddon) Response(f *px.Flow) {
	logger := configLoggerFieldsWithFlow(aud.logger, f)

	if aud.closed.Load() {
		logger.Warn("APIAuditor is being closed, not processing request")
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
			logger.Debug(
				"skipping accounting for unsupported API",
				"hostname", reqHostname,
			)
			return
		}

		// Only account when receiving good response codes
		_, shouldAccount := cacheOnlyResponseCodes[f.Response.StatusCode]
		if !shouldAccount {
			logger.Debug(
				"skipping accounting for non-200 response",
				"StatusCode", f.Response.StatusCode,
			)
			return
		}

		// convert the request to an internal TrafficObject
		reqAdapter := mitm.NewProxyRequestAdapter(f.Request) // generic wrapper for the mitm request

		tObjReq, err := schema.NewProxyRequest(reqAdapter, config.NewHeaderFilterGroup("empty", []string{}, []string{}))
		if err != nil {
			logger.Error("error creating TrafficObject from request", "error", err)
			return
		}

		// convert the response to an internal TrafficObject
		respAdapter := mitm.NewProxyResponseAdapter(f.Response) // generic wrapper for the mitm response

		tObjResp, err := schema.NewProxyResponse(respAdapter, config.NewHeaderFilterGroup("empty", []string{}, []string{}))
		if err != nil {
			logger.Error("error creating TrafficObject from response", "error", err)
			return
		}

		// account the cost
		auditOutput, err := aud.costCounter.Add(*tObjReq, *tObjResp)
		if err != nil {
			logger.Error("unable to create audit output", "error", err)
			return
		}

		// show the transaction
		aud.auditLogger.Info(
			"Transaction Received",
			"URL", auditOutput.URL,
			"Model", auditOutput.Model,
			"InputCost", auditOutput.InputCost,
			"OutputCost", auditOutput.OutputCost,
			"TotalReqCost", auditOutput.TotalReqCost,
			"SessionTotal", auditOutput.GrandTotal,
		)
	}()
}

func (aud *APIAuditorAddon) Close() error {
	if !aud.closed.Swap(true) {
		aud.logger.Debug("Closing...")
		aud.wg.Wait()
	}

	return nil
}

func NewAPIAuditor(logger *slog.Logger) *APIAuditorAddon {
	aud := &APIAuditorAddon{
		costCounter: schema.NewCostCounterDefaults(),
		logger:      logger.WithGroup("addons.APIAuditorAddon"),
		auditLogger: logger.WithGroup("API_Auditor"),
	}
	aud.closed.Store(false) // initialize as open
	return aud
}
