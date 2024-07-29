package writers

import (
	"log/slog"

	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
)

type ToStdOut struct {
	logger *slog.Logger
}

func (t *ToStdOut) Write(identifier string, bytes []byte) (int, error) {
	t.logger.Info(string(bytes), "identifier", identifier)
	return len(bytes), nil
}

func (t *ToStdOut) String() string {
	return "ToStdOut"
}

func newToStdOut(logger *slog.Logger) (*ToStdOut, error) {
	return &ToStdOut{
		logger: logger.With("writer", "ToStdOut"),
	}, nil
}

func NewToStdOut(logger *slog.Logger, _ string, _ formatters.MegaDumpFormatter) (MegaDumpWriter, error) {
	return newToStdOut(logger)
}
