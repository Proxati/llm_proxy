package transformers

import (
	"context"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
)

type Provider interface {
	HealthCheck() error
	Transform(
		ctx context.Context,
		oldReq *schema.ProxyRequest,
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
