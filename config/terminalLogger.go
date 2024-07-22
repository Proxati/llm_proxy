package config

import (
	"fmt"
	"log/slog"
	"os"
)

// terminalLogger controls the logging output to the terminal while the proxy is running
type terminalLogger struct {
	Verbose            bool   // if true, print runtime activity to stderr
	Debug              bool   // if true, print debug information to stderr
	Trace              bool   // if true, print detailed report caller tracing to stderr, for debugging
	logLevelHasBeenSet bool   // internal flag to track if the log level has been set
	sLoggerFormat      string // JSON or TXT ?
	slogHandlerOpts    *slog.HandlerOptions
}

func (tLo *terminalLogger) setupLoggerFormat() (logger *slog.Logger) {
	switch tLo.sLoggerFormat {
	case "json":
		fmt.Println("JSON log format")
		logger = slog.New(slog.NewJSONHandler(os.Stdout, tLo.slogHandlerOpts))
	default:
		fmt.Println("TXT log format")
		logger = slog.New(slog.NewTextHandler(os.Stdout, tLo.slogHandlerOpts))
	}
	return logger
}

// setLoggerLevel sets the log level based on verbose/debug values in the config object
func (tLo *terminalLogger) setLoggerLevel() {
	tLo.slogHandlerOpts = &slog.HandlerOptions{}
	if tLo.Debug {
		tLo.slogHandlerOpts.Level = slog.LevelDebug
		if tLo.Trace {
			tLo.slogHandlerOpts.AddSource = true
		}
	} else if tLo.Verbose {
		tLo.slogHandlerOpts.Level = slog.LevelInfo
	} else {
		tLo.slogHandlerOpts.Level = slog.LevelWarn
	}

	logger := tLo.setupLoggerFormat()
	slog.SetDefault(logger)
	slog.Debug("Global logger setup completed", "sLogLevel", tLo.slogHandlerOpts.Level)
	tLo.logLevelHasBeenSet = true
}

// getDebugLevel returns 1 if the log level is debug, 0 otherwise, for use in the proxy package
func (tLo *terminalLogger) getDebugLevel() int {
	if !tLo.logLevelHasBeenSet {
		tLo.setLoggerLevel()
	}
	switch tLo.slogHandlerOpts.Level {
	case slog.LevelDebug:
		return 1
	default:
		return 0
	}
}
