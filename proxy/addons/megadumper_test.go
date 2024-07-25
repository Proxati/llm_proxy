package addons

import (
	"log/slog"
	"os"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	md "github.com/proxati/llm_proxy/v2/proxy/addons/megadumper"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMegaDumpAddon(t *testing.T) {
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
		assert.Len(t, mda.writers, 1)
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
		assert.Len(t, mda.writers, 1)
	})
}

func TestNewLogDestinations(t *testing.T) {
	t.Run("Empty log target", func(t *testing.T) {
		logTarget := ""
		expectedDestinations := []md.LogDestination{md.WriteToStdOut}

		destinations, err := newLogDestinations(logTarget)

		assert.NoError(t, err)
		assert.Equal(t, expectedDestinations, destinations)
	})

	t.Run("Valid target", func(t *testing.T) {
		// Create a temporary directory
		logTarget, err := os.MkdirTemp("", "logDestTest")
		require.NoError(t, err)
		defer os.RemoveAll(logTarget)

		expectedDestinations := []md.LogDestination{md.WriteToDir}
		destinations, err := newLogDestinations(logTarget)
		require.NoError(t, err)
		assert.Equal(t, expectedDestinations, destinations)

		/*
			// delete and try again, should fail
			err = os.RemoveAll(logTarget)
			require.NoError(t, err)

			destinations, err = newLogDestinations(logTarget)
			assert.Error(t, err)
			assert.Nil(t, destinations)
		*/
	})
}

func TestNewWriters(t *testing.T) {
	t.Run("Empty destinations", func(t *testing.T) {
		logDestinations := []md.LogDestination{}
		logTarget := "/tmp/logs"
		formatter := &formatters.JSON{} // Assuming JSON formatter for simplicity

		writers, err := newWriters(logDestinations, logTarget, formatter)

		assert.NoError(t, err)
		assert.Empty(t, writers)
	})

	t.Run("StdOut destination", func(t *testing.T) {
		logDestinations := []md.LogDestination{md.WriteToStdOut}
		logTarget := ""
		formatter := &formatters.JSON{}

		writers, err := newWriters(logDestinations, logTarget, formatter)

		assert.NoError(t, err)
		assert.Len(t, writers, 1)
		// assert.IsType(t, &writers.StdOutWriter{}, writers[0])
	})

	t.Run("Dir destination", func(t *testing.T) {
		logDestinations := []md.LogDestination{md.WriteToDir}
		logTarget, err := os.MkdirTemp("", "logDestTest")
		require.NoError(t, err)
		defer os.RemoveAll(logTarget)
		formatter := &formatters.JSON{}

		writers, err := newWriters(logDestinations, logTarget, formatter)

		assert.NoError(t, err)
		assert.Len(t, writers, 1)
		// assert.IsType(t, &writers.DirWriter{}, writers[0])
	})

	t.Run("Invalid destination", func(t *testing.T) {
		logDestinations := []md.LogDestination{999} // Invalid destination
		logTarget := "/tmp/logs"
		formatter := &formatters.JSON{}

		writers, err := newWriters(logDestinations, logTarget, formatter)

		assert.Error(t, err)
		assert.Nil(t, writers)
	})

	t.Run("Multiple destinations", func(t *testing.T) {
		logDestinations := []md.LogDestination{md.WriteToStdOut, md.WriteToDir}
		logTarget, err := os.MkdirTemp("", "logDestTest")
		require.NoError(t, err)
		defer os.RemoveAll(logTarget)
		formatter := &formatters.JSON{}

		writers, err := newWriters(logDestinations, logTarget, formatter)

		assert.NoError(t, err)
		assert.Len(t, writers, 2)
		// assert.IsType(t, &writers.StdOutWriter{}, writers[0])
		// assert.IsType(t, &writers.DirWriter{}, writers[1])
	})

}
