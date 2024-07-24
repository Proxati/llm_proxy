package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_SetLoggerLevel_ImplicitCallToSetLoggerLevel(t *testing.T) {
	cfg := &Config{}

	assert.Nil(t, cfg.terminalLogger)

	assert.Equal(t, 0, cfg.IsDebugEnabled())
	assert.Equal(t, true, cfg.terminalLogger.logLevelHasBeenSet)
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	assert.IsType(t, &Config{}, cfg)
}
