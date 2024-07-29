package addons

import (
	"log/slog"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewMegaDumpAddon(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()

	t.Run("JSON", func(t *testing.T) {
		logTarget := "/tmp/logs"
		logFormat := config.LogFormat_JSON
		logSources := config.LogSourceConfig{}
		filterReqHeaders := []string{}
		filterRespHeaders := []string{}

		mda, err := NewMegaTrafficDumperAddon(testLogger, logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)

		assert.NoError(t, err)
		assert.NotNil(t, mda)
		assert.Equal(t, logSources, mda.logSources)
		assert.Len(t, mda.logDestinationConfigs, 1)
	})

	t.Run("TXT", func(t *testing.T) {
		logTarget := "/tmp/logs"
		logFormat := config.LogFormat_TXT
		logSources := config.LogSourceConfig{}
		filterReqHeaders := []string{}
		filterRespHeaders := []string{}

		mda, err := NewMegaTrafficDumperAddon(testLogger, logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)

		assert.NoError(t, err)
		assert.NotNil(t, mda)
		assert.Equal(t, logSources, mda.logSources)
		assert.Len(t, mda.logDestinationConfigs, 1)
	})

	t.Run("< empty >", func(t *testing.T) {
		logTarget := ""
		logFormat := config.LogFormat_TXT
		logSources := config.LogSourceConfig{}
		filterReqHeaders := []string{}
		filterRespHeaders := []string{}

		mda, err := NewMegaTrafficDumperAddon(testLogger, logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)

		assert.NoError(t, err)
		assert.NotNil(t, mda)
		assert.Equal(t, logSources, mda.logSources)
		assert.Len(t, mda.logDestinationConfigs, 1)
	})
}
