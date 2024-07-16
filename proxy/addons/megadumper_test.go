package addons

import (
	"os"
	"testing"

	"github.com/proxati/llm_proxy/config"
	md "github.com/proxati/llm_proxy/proxy/addons/megadumper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMegaDirDumper_JSON_LogDir(t *testing.T) {
	logTarget := "/tmp/logs"
	logFormat := config.TrafficLog_JSON
	logSources := config.LogSourceConfig{}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaDirDumper(logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)

	assert.NoError(t, err)
	assert.NotNil(t, mda)
	assert.Equal(t, logSources, mda.logSources)
	assert.Len(t, mda.writers, 1)
}

func TestNewMegaDirDumper_TXT_LOGFILE(t *testing.T) {
	logTarget := "/tmp/logs"
	logFormat := config.TrafficLog_TXT
	logSources := config.LogSourceConfig{}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaDirDumper(logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)

	assert.NoError(t, err)
	assert.NotNil(t, mda)
	assert.Equal(t, logSources, mda.logSources)
	assert.Len(t, mda.writers, 1)
}

func TestNewLogDestinationsEmptyLogTarget(t *testing.T) {
	logTarget := ""
	expectedDestinations := []md.LogDestination{md.WriteToStdOut}

	destinations, err := newLogDestinations(logTarget)

	assert.NoError(t, err)
	assert.Equal(t, expectedDestinations, destinations)
}

func TestNewLogDestinationsValidLogTarget(t *testing.T) {
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
}
