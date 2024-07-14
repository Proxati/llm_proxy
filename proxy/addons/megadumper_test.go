package addons

import (
	"testing"

	"github.com/proxati/llm_proxy/config"
	"github.com/stretchr/testify/assert"
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
