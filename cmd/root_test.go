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
			t.Parallel()
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
		{"LevelDebug_TXT_Terminal", slog.LevelDebug, config.LogFormatTXT, true, splashText},
		{"LevelDebug_TXT_NonTerminal", slog.LevelDebug, config.LogFormatTXT, false, ""},
		{"LevelDebug_JSON_Terminal", slog.LevelDebug, config.LogFormatJSON, true, ""},
		{"LevelDebug_JSON_NonTerminal", slog.LevelDebug, config.LogFormatJSON, false, ""},

		{"LevelInfo_TXT_Terminal", slog.LevelInfo, config.LogFormatTXT, true, splashText},
		{"LevelInfo_TXT_NonTerminal", slog.LevelInfo, config.LogFormatTXT, false, ""},
		{"LevelInfo_JSON_Terminal", slog.LevelInfo, config.LogFormatJSON, true, ""},
		{"LevelInfo_JSON_NonTerminal", slog.LevelInfo, config.LogFormatJSON, false, ""},

		{"LevelWarn_TXT_Terminal", slog.LevelWarn, config.LogFormatTXT, true, ""},
		{"LevelWarn_TXT_NonTerminal", slog.LevelWarn, config.LogFormatTXT, false, ""},
		{"LevelWarn_JSON_Terminal", slog.LevelWarn, config.LogFormatJSON, true, ""},
		{"LevelWarn_JSON_NonTerminal", slog.LevelWarn, config.LogFormatJSON, false, ""},

		{"LevelError_TXT_Terminal", slog.LevelError, config.LogFormatTXT, true, ""},
		{"LevelError_TXT_NonTerminal", slog.LevelError, config.LogFormatTXT, false, ""},
		{"LevelError_JSON_Terminal", slog.LevelError, config.LogFormatJSON, true, ""},
		{"LevelError_JSON_NonTerminal", slog.LevelError, config.LogFormatJSON, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()
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
		{"ValidFormats_TXT_JSON", "txt", "json", config.LogFormatTXT, config.LogFormatJSON, false},
		{"ValidFormats_JSON_TXT", "json", "txt", config.LogFormatJSON, config.LogFormatTXT, false},
		{"InvalidTerminalFormat", "invalid", "json", config.LogFormatTXT, config.LogFormatJSON, true},
		{"InvalidTrafficFormat", "txt", "invalid", config.LogFormatTXT, config.LogFormatJSON, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.NewDefaultConfig()
			logFmt, err := setupLogFormats(cfg, tt.terminalLogFormat, tt.trafficLogFormat)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, config.LogFormatTXT, cfg.GetTerminalOutputFormat())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTermFmt, cfg.GetTerminalOutputFormat())
				assert.Equal(t, tt.expectedTrafficFmt, cfg.TrafficLogger.LogFmt)
				assert.Equal(t, tt.expectedTermFmt, logFmt)
			}
		})
	}
}
