package writers

import (
	"log/slog"

	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/writers/remote/rest"
)

const DefaultRestTimeout = 5

type ToAsyncRest struct {
	endpoint  rest.Endpoint
	target    string
	formatter formatters.MegaDumpFormatter
	logger    *slog.Logger
}

func (t *ToAsyncRest) Write(identifier string, data []byte) (int, error) {
	err := t.endpoint.POST(identifier, data)
	if err != nil {
		return 0, err
	}
	t.logger.Info("Successfully sent data", "identifier", identifier, "endpoint", t.endpoint.String())
	return len(data), nil
}

func (t *ToAsyncRest) String() string {
	return "ToAsyncRest: " + t.target
}

func newToAsyncREST(logger *slog.Logger, target string, formatter formatters.MegaDumpFormatter) (*ToAsyncRest, error) {
	logger = logger.WithGroup("ToAsyncRest").With("target", target, "formatter", formatter)

	return &ToAsyncRest{
		endpoint:  rest.NewEndpointSyncREST(logger, "ToAsyncRest", target),
		target:    target,
		formatter: formatter,
		logger:    logger,
	}, nil
}

func NewToAsyncREST(logger *slog.Logger, target string, formatter formatters.MegaDumpFormatter) (MegaDumpWriter, error) {
	return newToAsyncREST(logger, target, formatter)
}
