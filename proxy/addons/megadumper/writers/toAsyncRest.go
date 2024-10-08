package writers

import (
	"log/slog"

	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/writers/remote/rest"
)

// ToAsyncRest is a writer that sends data to a remote REST endpoint
type ToAsyncRest struct {
	endpoint  rest.Endpoint
	target    string
	formatter formatters.MegaDumpFormatter
	logger    *slog.Logger
}

// NewToAsyncREST creates a new ToAsyncRest writer object
// Parameters:
// - logger: a slog.Logger object
// - target: the target URL to send the data to
// - formatter: a formatters.MegaDumpFormatter object, probably a JSON formatter
func NewToAsyncREST(
	logger *slog.Logger,
	target string,
	formatter formatters.MegaDumpFormatter,
) (*ToAsyncRest, error) {
	logger = logger.WithGroup("ToAsyncRest").With("target", target, "formatter", formatter)

	return &ToAsyncRest{
		endpoint:  rest.NewEndpointSyncREST(logger, "ToAsyncRest", target),
		target:    target,
		formatter: formatter,
		logger:    logger,
	}, nil
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
