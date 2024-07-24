package config

import "log/slog"

const (
	DefaultListenAddr = "127.0.0.1:8080"
	DefaultCacheDir   = "/tmp/llm_proxy"
)

// Config is the main config mega-struct
type Config struct {
	*httpBehavior
	*trafficLogger
	AppMode        AppMode
	Cache          *cacheBehavior
	terminalLogger *terminalLogger
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

func (cfg *Config) GetLogger() *slog.Logger {
	if cfg.terminalLogger.logger == nil {
		cfg.SetLoggerLevel()
	}
	return cfg.terminalLogger.logger
}

func (cfg *Config) EnableOutputDebug() {
	tlo := cfg.getTerminalLogger()
	tlo.Verbose = false
	tlo.Debug = true
	tlo.logLevelHasBeenSet = false
	cfg.SetLoggerLevel()
}

func (cfg *Config) EnableOutputVerbose() {
	tlo := cfg.getTerminalLogger()
	tlo.Verbose = true
	tlo.Debug = false
	tlo.logLevelHasBeenSet = false
	cfg.SetLoggerLevel()
}

func (cfg *Config) EnableOutputTrace() {
	tlo := cfg.getTerminalLogger()
	tlo.Verbose = false
	tlo.Debug = true
	tlo.Trace = true
	tlo.logLevelHasBeenSet = false
	cfg.SetLoggerLevel()
}

func (cfg *Config) SetTerminalOutputFormat(terminalLogFormat string) (LogFormat, error) {
	tlo := cfg.getTerminalLogger()
	var err error
	tlo.TerminalSloggerFormat, err = StringToLogFormat(terminalLogFormat)
	tlo.logLevelHasBeenSet = false
	cfg.SetLoggerLevel()

	if err != nil {
		return 0, err
	}
	return tlo.TerminalSloggerFormat, nil
}

func NewDefaultConfig() *Config {
	return &Config{
		httpBehavior: &httpBehavior{
			Listen:                DefaultListenAddr,
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
			Dir: DefaultCacheDir,
			TTL: 0,
		},
	}
}
