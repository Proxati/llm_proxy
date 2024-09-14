package transformers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/bulkhead"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
)

// configRetryPolicy create a RetryPolicy that only handles 500 responses, with backoff delays between retries
func configRetryPolicy(logger *slog.Logger, retryCount int, initialRetryTime, maxRetryTime time.Duration) retrypolicy.RetryPolicy[*http.Response] {
	return retrypolicy.Builder[*http.Response]().
		HandleIf(func(response *http.Response, _ error) bool {
			// if upstream responds with a 500, retry
			return response != nil && response.StatusCode == 500
		}).
		WithMaxRetries(retryCount).
		WithBackoff(initialRetryTime, maxRetryTime).
		WithJitter(250 * time.Millisecond).
		OnRetryScheduled(func(e failsafe.ExecutionScheduledEvent[*http.Response]) {
			logger.Warn("Waiting for retry", "attempts", e.Attempts(), "delay", e.Delay)
		}).
		OnRetry(func(e failsafe.ExecutionEvent[*http.Response]) {
			logger.Warn("Retrying", "attempts", e.Attempts())
		}).
		OnAbort(func(e failsafe.ExecutionEvent[*http.Response]) {
			logger.Error("Aborting retry", "attempts", e.Attempts())
		}).
		OnRetriesExceeded(func(e failsafe.ExecutionEvent[*http.Response]) {
			logger.Error("Retries exceeded", "attempts", e.Attempts())
		}).Build()
}

type httpBundle struct {
	client      *http.Client
	retryPolicy retrypolicy.RetryPolicy[*http.Response]
	bulkhead    bulkhead.Bulkhead[any]
	url         string
}

func newHttpBundle(logger *slog.Logger, url string, concurrency, retryCount int, bulkHeadTimeout, httpClientTimeout, initialRetryTime, maxRetryTime time.Duration) *httpBundle {
	// Create a Bulkhead with a limit of N concurrent executions
	bh := bulkhead.Builder[any](uint(concurrency)).
		WithMaxWaitTime(bulkHeadTimeout).
		OnFull(func(e failsafe.ExecutionEvent[any]) {
			logger.Warn("Bulkhead full")
		}).
		Build()

	return &httpBundle{
		client:      &http.Client{Timeout: httpClientTimeout},
		retryPolicy: configRetryPolicy(logger, retryCount, initialRetryTime, maxRetryTime),
		bulkhead:    bh,
		url:         url,
	}
}

type HttpProvider struct {
	logger            *slog.Logger
	TransformerConfig *config.Transformer
	primaryBundle     *httpBundle
	healthCheckBundle *httpBundle
	context           context.Context
	cancelCtx         context.CancelFunc
}

// NewHttpProvider creates a new HttpProvider object with the given logger and transformer config
// Notes:
// Tcfg.Concurrency is the number of concurrent requests allowed
// Tcfg.RequestTimeout is the maximum time allowed for a request to begin
func NewHttpProvider(logger *slog.Logger, Tcfg *config.Transformer) (*HttpProvider, error) {
	if Tcfg == nil {
		return nil, errors.New("no transformer config provided")
	}
	logger = logger.WithGroup("HttpProvider").With("ServiceName", Tcfg.Name)

	healthCheckURL, err := Tcfg.GetHealthCheckURL()
	if err != nil {
		return nil, fmt.Errorf("unable to build health check URL: %w", err)
	}

	bhTimeout := Tcfg.RequestTimeout + 1*time.Second // this might be wrong
	healthCheckRetry := 3

	// TODO this is wrong
	hcInitialRetryTime := Tcfg.HealthCheck.Interval
	hcInitialRetryTimeMax := Tcfg.HealthCheck.Interval + 1*time.Second

	tfB := newHttpBundle(
		logger.WithGroup("Primary"), Tcfg.URL.String(), Tcfg.Concurrency, Tcfg.RetryCount, bhTimeout, Tcfg.RequestTimeout, Tcfg.InitialRetryTime, Tcfg.MaxRetryTime)
	hcB := newHttpBundle(
		logger.WithGroup("HealthCheck"), healthCheckURL, 1, healthCheckRetry, bhTimeout, Tcfg.HealthCheck.Timeout, hcInitialRetryTime, hcInitialRetryTimeMax)

	ctx, cancel := context.WithCancel(context.Background())
	return &HttpProvider{
		logger:            logger,
		TransformerConfig: Tcfg,
		primaryBundle:     tfB,
		healthCheckBundle: hcB,
		context:           ctx,
		cancelCtx:         cancel,
	}, nil
}

func (ht *HttpProvider) String() string {
	return fmt.Sprintf("HttpProvider{Name: %s, URL: %s}", ht.TransformerConfig.Name, ht.TransformerConfig.URL.String())
}

func (ht *HttpProvider) HealthCheck() error {
	// Create a new request
	req, err := http.NewRequest("GET", ht.healthCheckBundle.url, nil)
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	// Send the request
	resp, err := ht.healthCheckBundle.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %s", resp.Status)
	}

	return nil
}

func (ht *HttpProvider) Transform(
	ctx context.Context,
	flow proxyadapters.FlowReaderAdapter,
	req proxyadapters.ResponseReaderAdapter, resp proxyadapters.RequestReaderAdapter,
) (proxyadapters.ResponseReaderAdapter, proxyadapters.RequestReaderAdapter, error) {
	logger := ht.logger.With("flow", flow, "request", req, "response", resp)
	logger.Debug("Transforming")
	return nil, nil, nil
}
