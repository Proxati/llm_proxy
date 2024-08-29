package writers

import (
	"log/slog"

	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
)

// ToStdOut is a very basic writer that writes to stdout
type ToStdOut struct {
	logger *slog.Logger
}

// NewToStdOut creates a new ToStdOut writer object
func NewToStdOut(logger *slog.Logger, _ string, _ formatters.MegaDumpFormatter) (*ToStdOut, error) {
	return &ToStdOut{
		logger: logger.With("writer", "ToStdOut"),
	}, nil
}

// Write writes the given bytes to standard out
func (t *ToStdOut) Write(identifier string, bytes []byte) (int, error) {
	t.logger.Info(string(bytes), "identifier", identifier)
	return len(bytes), nil
}

// String returns the human-readable name of this writer
func (t *ToStdOut) String() string {
	return "ToStdOut"
}
