package cmd

import (
	"log/slog"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestSetupTerminalOutputLevel(t *testing.T) {
	tests := []struct {
		name        string
		debugMode   bool
		verboseMode bool
		traceMode   bool
		expected    string
	}{
		{"DebugMode", true, false, false, "debug"},
		{"VerboseMode", false, true, false, "verbose"},
		{"TraceMode", true, false, true, "trace"},
		{"DefaultMode", false, false, false, "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			cfg.SetLoggerLevel() // two constructors for the logger is annoying

			setupTerminalOutputLevel(cfg, tt.debugMode, tt.verboseMode, tt.traceMode)
			if tt.debugMode {
				assert.Equal(t, slog.LevelDebug, cfg.GetLoggerLevel())
				assert.Equal(t, tt.traceMode, cfg.IsTraceEnabled())
			} else if tt.verboseMode {
				assert.Equal(t, slog.LevelInfo, cfg.GetLoggerLevel())
				assert.False(t, cfg.IsTraceEnabled())
			} else {
				assert.Equal(t, slog.LevelWarn, cfg.GetLoggerLevel())
				assert.False(t, cfg.IsTraceEnabled())
			}
		})
	}
}

func TestPrintSplash(t *testing.T) {
	splashText := "testing"
	tests := []struct {
		testName       string
		logLvl         slog.Level
		logFmt         config.LogFormat
		isTerminal     bool
		expectedOutput string
	}{
		{"LevelDebug_TXT_Terminal", slog.LevelDebug, config.LogFormat_TXT, true, splashText},
		{"LevelDebug_TXT_NonTerminal", slog.LevelDebug, config.LogFormat_TXT, false, ""},
		{"LevelDebug_JSON_Terminal", slog.LevelDebug, config.LogFormat_JSON, true, ""},
		{"LevelDebug_JSON_NonTerminal", slog.LevelDebug, config.LogFormat_JSON, false, ""},

		{"LevelInfo_TXT_Terminal", slog.LevelInfo, config.LogFormat_TXT, true, splashText},
		{"LevelInfo_TXT_NonTerminal", slog.LevelInfo, config.LogFormat_TXT, false, ""},
		{"LevelInfo_JSON_Terminal", slog.LevelInfo, config.LogFormat_JSON, true, ""},
		{"LevelInfo_JSON_NonTerminal", slog.LevelInfo, config.LogFormat_JSON, false, ""},

		{"LevelWarn_TXT_Terminal", slog.LevelWarn, config.LogFormat_TXT, true, ""},
		{"LevelWarn_TXT_NonTerminal", slog.LevelWarn, config.LogFormat_TXT, false, ""},
		{"LevelWarn_JSON_Terminal", slog.LevelWarn, config.LogFormat_JSON, true, ""},
		{"LevelWarn_JSON_NonTerminal", slog.LevelWarn, config.LogFormat_JSON, false, ""},

		{"LevelError_TXT_Terminal", slog.LevelError, config.LogFormat_TXT, true, ""},
		{"LevelError_TXT_NonTerminal", slog.LevelError, config.LogFormat_TXT, false, ""},
		{"LevelError_JSON_Terminal", slog.LevelError, config.LogFormat_JSON, true, ""},
		{"LevelError_JSON_NonTerminal", slog.LevelError, config.LogFormat_JSON, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			sp := printSplash(tt.logLvl, tt.logFmt, tt.isTerminal, splashText)
			assert.Equal(t, tt.expectedOutput, sp)
		})
	}
}

func TestSetupLogFormats(t *testing.T) {
	tests := []struct {
		name               string
		terminalLogFormat  string
		trafficLogFormat   string
		expectedTermFmt    config.LogFormat
		expectedTrafficFmt config.LogFormat
		expectError        bool
	}{
		{"ValidFormats_TXT_JSON", "txt", "json", config.LogFormat_TXT, config.LogFormat_JSON, false},
		{"ValidFormats_JSON_TXT", "json", "txt", config.LogFormat_JSON, config.LogFormat_TXT, false},
		{"InvalidTerminalFormat", "invalid", "json", config.LogFormat_TXT, config.LogFormat_JSON, true},
		{"InvalidTrafficFormat", "txt", "invalid", config.LogFormat_TXT, config.LogFormat_JSON, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			logFmt, err := setupLogFormats(cfg, tt.terminalLogFormat, tt.trafficLogFormat)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTermFmt, cfg.GetTerminalOutputFormat())
				assert.Equal(t, tt.expectedTrafficFmt, cfg.TrafficLogFmt)
				assert.Equal(t, tt.expectedTermFmt, logFmt)
			}
		})
	}
}
