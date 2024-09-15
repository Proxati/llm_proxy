package transformers

import (
	"context"
	"log/slog"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
)

type Provider interface {
	HealthCheck(context.Context) error
	Transform(
		ctx context.Context,
		logger *slog.Logger,
		oldReq *schema.ProxyRequest,
		oldResp *schema.ProxyResponse,
		newReq *schema.ProxyRequest,
		newResp *schema.ProxyResponse,
	) (
		*schema.ProxyRequest, // new request, possibly modified or nil
		*schema.ProxyResponse, // new response, possibly modified or nil
		error,
	)
	Close() error
	GetTransformerConfig() config.Transformer
}
