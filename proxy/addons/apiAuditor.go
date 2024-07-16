package addons

import (
	"fmt"
	"sync"
	"sync/atomic"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/proxati/llm_proxy/schema"
	"github.com/proxati/llm_proxy/schema/providers"
	log "github.com/sirupsen/logrus"
)

// APIAuditorAddon log connection and flow
type APIAuditorAddon struct {
	px.BaseAddon
	costCounter *schema.CostCounter
	closed      atomic.Bool
	wg          sync.WaitGroup
}

func (aud *APIAuditorAddon) Response(f *px.Flow) {
	if aud.closed.Load() {
		log.Warn("APIAuditor is being closed, not processing request")
		return
	}
	if f.Response == nil {
		log.Debugf("skipping accounting for nil response: %s", f.Request.URL)
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
			log.Debugf("skipping accounting for unsupported API: %s", reqHostname)
			return
		}

		// Only account when receiving good response codes
		_, shouldAccount := cacheOnlyResponseCodes[f.Response.StatusCode]
		if !shouldAccount {
			log.Debugf("skipping accounting for non-200 response: %s", f.Request.URL)
			return
		}

		// convert the request to an internal TrafficObject
		tObjReq, err := schema.NewProxyRequestFromMITMRequest(f.Request, []string{})
		if err != nil {
			log.Errorf("error creating TrafficObject from request: %s", f.Request.URL)
			return
		}

		// convert the response to an internal TrafficObject
		tObjResp, err := schema.NewProxyResponseFromMITMResponse(f.Response, []string{})
		if err != nil {
			log.Errorf("error creating TrafficObject from response: %s", err)
			return
		}

		// account the cost, TODO: returns what?
		auditOutput, err := aud.costCounter.Add(*tObjReq, *tObjResp)
		if err != nil {
			log.Errorf("error accounting response: %s", err)
		}
		fmt.Println(auditOutput)
	}()
}

func (aud *APIAuditorAddon) Close() error {
	if !aud.closed.Swap(true) {
		log.Debug("Waiting for APIAuditor shutdown...")
		aud.wg.Wait()
	}

	return nil
}

func NewAPIAuditor() *APIAuditorAddon {
	aud := &APIAuditorAddon{
		costCounter: schema.NewCostCounterDefaults(),
	}
	aud.closed.Store(false) // initialize as open
	return aud
}
