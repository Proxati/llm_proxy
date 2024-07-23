package config

import "log/slog"

// Config is the main config mega-struct
type Config struct {
	AppMode AppMode
	*httpBehavior
	*terminalLogger
	*trafficLogger
	Cache *cacheBehavior
}

func (cfg *Config) getTerminalLogger() *terminalLogger {
	if cfg.terminalLogger == nil {
		cfg.terminalLogger = &terminalLogger{}
	}
	return cfg.terminalLogger
}

func (cfg *Config) SetLoggerLevel() {
	cfg.getTerminalLogger().setLoggerLevel()
}

func (cfg *Config) IsDebugEnabled() int {
	return cfg.getTerminalLogger().getDebugLevel()
}

// IsVerboseOrHigher returns 1 if the log level is verbose or higher
func (cfg *Config) IsVerboseOrHigher() bool {
	if cfg.getTerminalLogger().Verbose || cfg.getTerminalLogger().Debug || cfg.getTerminalLogger().Trace {
		return true
	}
	return false
}

func NewDefaultConfig() *Config {
	return &Config{
		httpBehavior: &httpBehavior{
			Listen:                "127.0.0.1:8080",
			CertDir:               "",
			InsecureSkipVerifyTLS: false,
			NoHttpUpgrader:        false,
		},
		terminalLogger: &terminalLogger{
			Verbose:               false,
			Debug:                 false,
			Trace:                 false,
			logLevelHasBeenSet:    false,
			TerminalSloggerFormat: LogFormat_TXT,
			slogHandlerOpts:       &slog.HandlerOptions{},
		},
		trafficLogger: &trafficLogger{
			OutputDir:         "",
			TrafficLogFmt:     LogFormat_JSON,
			FilterReqHeaders:  append([]string{}, defaultFilterHeaders...), // append empty to deep copy the source slice
			FilterRespHeaders: append([]string{}, defaultFilterHeaders...),
		},
		Cache: &cacheBehavior{
			Dir: "/tmp/llm_proxy",
			TTL: 0,
		},
	}
}
