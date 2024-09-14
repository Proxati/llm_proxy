package transformers

import (
	"context"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
)

type Provider interface {
	HealthCheck() error
	Transform(
		ctx context.Context,
		oldReq proxyadapters.RequestReaderAdapter,
		newReq proxyadapters.RequestReaderAdapter,
		newResp proxyadapters.ResponseReaderAdapter,
	) (
		proxyadapters.RequestReaderAdapter, // new request, possibly modified or nil
		proxyadapters.ResponseReaderAdapter, // new response, possibly modified or nil
		error,
	)
}
