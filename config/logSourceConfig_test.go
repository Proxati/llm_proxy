package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogSourceConfig_String(t *testing.T) {
	t.Parallel()
	config := LogSourceConfig{
		LogConnectionStats: true,
		LogRequestHeaders:  true,
		LogRequest:         true,
		LogResponseHeaders: true,
		LogResponse:        true,
	}

	expected := `{"LogConnectionStats":true,"LogRequestHeaders":true,"LogRequest":true,"LogResponseHeaders":true,"LogResponse":true}`
	assert.Equal(t, expected, config.String())
}
