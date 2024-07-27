package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_SetLoggerLevel_Debug(t *testing.T) {
	t.Parallel()
	cfg := &terminalLogger{
		Debug: true,
	}

	assert.Equal(t, false, cfg.logLevelHasBeenSet)
	cfg.setLoggerLevel()

	assert.Equal(t, 1, cfg.getDebugLevel())
	assert.Equal(t, true, cfg.logLevelHasBeenSet)
}

func TestConfig_SetLoggerLevel_Verbose(t *testing.T) {
	t.Parallel()
	cfg := &terminalLogger{
		Verbose: true,
	}

	assert.Equal(t, false, cfg.logLevelHasBeenSet)
	cfg.setLoggerLevel()

	assert.Equal(t, 0, cfg.getDebugLevel())
	assert.Equal(t, true, cfg.logLevelHasBeenSet)
}

func TestConfig_SetLoggerLevel_Default(t *testing.T) {
	t.Parallel()
	cfg := &terminalLogger{}

	assert.Equal(t, false, cfg.logLevelHasBeenSet)
	cfg.setLoggerLevel()

	assert.Equal(t, 0, cfg.getDebugLevel())
	assert.Equal(t, true, cfg.logLevelHasBeenSet)
}
