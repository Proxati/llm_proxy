package addons

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStdOutLogger(t *testing.T) {
	logger := NewStdOutLogger(slog.Default())
	assert.NotNil(t, logger)
}
