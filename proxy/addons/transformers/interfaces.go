package transformers

import (
	"context"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
)

type Provider interface {
	HealthCheck() error
	Transform(context.Context, proxyadapters.FlowReaderAdapter, proxyadapters.ResponseReaderAdapter, proxyadapters.RequestReaderAdapter) (proxyadapters.ResponseReaderAdapter, proxyadapters.RequestReaderAdapter, error)
}
