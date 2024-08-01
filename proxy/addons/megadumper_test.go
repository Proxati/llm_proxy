package addons

import (
	"log/slog"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	px "github.com/proxati/mitmproxy/proxy"
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

func TestMegaTrafficDumper_Requestheaders_NilFlow(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()

	logTarget := "/tmp/logs"
	logFormat := config.LogFormat_JSON
	logSources := config.LogSourceConfig{}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaTrafficDumperAddon(testLogger, logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)
	assert.NoError(t, err)
	assert.NotNil(t, mda)

	t.Run("NilFlow", func(t *testing.T) {
		flow := &px.Flow{
			Request:  nil,
			Response: nil,
		}

		assert.NotPanics(t, func() {
			mda.Requestheaders(flow)
		})
	})
}

func TestMegaTrafficDumper_Requestheaders_ValidFlow(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()

	logTarget := "/tmp/logs"
	logFormat := config.LogFormat_JSON
	logSources := config.LogSourceConfig{}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaTrafficDumperAddon(testLogger, logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)
	assert.NoError(t, err)
	assert.NotNil(t, mda)

	t.Run("ValidFlow", func(t *testing.T) {
		flow := &px.Flow{
			Request:  &px.Request{},
			Response: &px.Response{},
		}

		assert.NotPanics(t, func() {
			mda.Requestheaders(flow)
		})
	})
}

func TestMegaTrafficDumper_Close(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()

	logTarget := "/tmp/logs"
	logFormat := config.LogFormat_JSON
	logSources := config.LogSourceConfig{}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaTrafficDumperAddon(testLogger, logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)
	assert.NoError(t, err)
	assert.NotNil(t, mda)

	t.Run("Close", func(t *testing.T) {
		assert.NoError(t, mda.Close())
		assert.True(t, mda.closed.Load())
	})
}

// unable to get this test working, because .Done() isn't working as expected
/*
func TestMegaTrafficDumper_LogWriting(t *testing.T) {
	t.Parallel()
	w := os.Stdout
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	testLogger := slog.New(slog.NewTextHandler(w, handlerOpts))

	logTarget := t.TempDir()
	logFormat := config.LogFormat_JSON
	logSources := config.LogSourceConfigAllTrue
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaTrafficDumperAddon(
		testLogger, logTarget, logFormat, logSources, filterReqHeaders, filterRespHeaders)
	assert.NoError(t, err)
	assert.NotNil(t, mda)

	testReq := &px.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "example.com",
			Path:   "/flow",
		},
		Proto:  "HTTP/1.1",
		Header: map[string][]string{"User-Agent": {"TestAgent"}},
		Body:   []byte(`{"flow":"data"}`),
	}

	t.Run("LogWriting", func(t *testing.T) {
		flow := &px.Flow{
			Request: testReq,
		}

		watch, err := fileUtils.NewFileWatcher(testLogger, logTarget)
		require.NoError(t, err)
		require.NotNil(t, watch)

		assert.NotPanics(t, func() {
			mda.Requestheaders(flow)
		})

		// this should error, because the log file is not until later
		err = fileUtils.WaitForFile(testLogger, watch, 100*time.Millisecond)
		require.Error(t, err)

		flow.Done()

		err = fileUtils.WaitForFile(testLogger, watch, 5*time.Second)
		require.Error(t, err)

		time.Sleep(2 * time.Second)

		// List the contents of the logTarget directory
		files, err := os.ReadDir(logTarget)
		require.NoError(t, err)
		for _, file := range files {
			testLogger.Info("Found file", "fileName", file.Name())
		}

		logFiles, err := filepath.Glob(filepath.Join(logTarget, "*.json"))
		require.NoError(t, err)
		require.Equal(t, 1, len(logFiles))
	})
}
*/
