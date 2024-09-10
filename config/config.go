package config

import "log/slog"

const (
	defaultListenAddr = "127.0.0.1:8080"
	defaultCacheDir   = "/tmp/llm_proxy"
)

// Config is the main config mega-struct
type Config struct {
	AppMode        AppMode
	Cache          *cacheBehavior
	HeaderFilters  *HeaderFiltersContainer
	HTTPBehavior   *httpBehavior
	TrafficLogger  *TrafficLogger
	terminalLogger *terminalLogger
}

func (cfg *Config) getTerminalLogger() *terminalLogger {
	if cfg.terminalLogger == nil {
		cfg.terminalLogger = &terminalLogger{}
	}
	return cfg.terminalLogger
}

// SetLoggerLevel is the external API to configure the logger with the specified level
func (cfg *Config) SetLoggerLevel() {
	cfg.getTerminalLogger().setLoggerLevel()
}

// GetLoggerLevel returns the current log level
func (cfg *Config) GetLoggerLevel() slog.Level {
	return cfg.getTerminalLogger().slogHandlerOpts.Level.Level()
}

// IsTraceEnabled returns true if the log is configured to add the source file to the log output
func (cfg *Config) IsTraceEnabled() bool {
	l := cfg.getTerminalLogger()
	return l.slogHandlerOpts.AddSource
}

// IsVerboseOrHigher returns 1 if the log level is verbose or higher
func (cfg *Config) IsVerboseOrHigher() bool {
	switch cfg.GetLoggerLevel() {
	case slog.LevelInfo, slog.LevelDebug:
		return true
	default:
		return false
	}
}

// GetLogger returns the slogger
func (cfg *Config) GetLogger() *slog.Logger {
	if cfg.getTerminalLogger().logger == nil {
		cfg.SetLoggerLevel()
	}
	return cfg.terminalLogger.logger
}

// EnableOutputDebug is the public API to enable debug output
func (cfg *Config) EnableOutputDebug() {
	tlo := cfg.getTerminalLogger()
	tlo.Verbose = false
	tlo.Debug = true
	tlo.logLevelHasBeenSet = false
	cfg.SetLoggerLevel()
}

// EnableOutputVerbose is the public API to enable verbose output
func (cfg *Config) EnableOutputVerbose() {
	tlo := cfg.getTerminalLogger()
	tlo.Verbose = true
	tlo.Debug = false
	tlo.logLevelHasBeenSet = false
	cfg.SetLoggerLevel()
}

// EnableOutputTrace is the public API to enable trace output
func (cfg *Config) EnableOutputTrace() {
	tlo := cfg.getTerminalLogger()
	tlo.Verbose = false
	tlo.Debug = true
	tlo.Trace = true
	tlo.logLevelHasBeenSet = false
	cfg.SetLoggerLevel()
}

// SetTerminalOutputFormat sets the terminal output format. It turns an unformatted string from
// user input into an enum, and will return an error if the input string is not supported.
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

// SetTrafficLogFormat sets the traffic log format. It turns an unformatted string from
// user input into an enum, and will return an error if the input string is not supported.
func (cfg *Config) SetTrafficLogFormat(logfmt string) error {
	var err error
	cfg.TrafficLogger.LogFmt, err = StringToLogFormat(logfmt)
	return err
}

// GetTerminalOutputFormat returns the current traffic log format enum
func (cfg *Config) GetTerminalOutputFormat() LogFormat {
	return cfg.getTerminalLogger().TerminalSloggerFormat
}

// NewDefaultConfig creates a new mega config object with all default values
func NewDefaultConfig() *Config {
	cb, err := newCacheBehavior(defaultCacheDir, CacheEngineBolt.String())
	if err != nil {
		// this should never happen!
		panic(err)
	}

	return &Config{
		HTTPBehavior: &httpBehavior{
			Listen:                defaultListenAddr,
			CertDir:               "",
			InsecureSkipVerifyTLS: false,
			NoHTTPUpgrader:        false,
		},
		terminalLogger: &terminalLogger{
			Verbose:               false,
			Debug:                 false,
			Trace:                 false,
			logLevelHasBeenSet:    false,
			TerminalSloggerFormat: LogFormatTXT,
			slogHandlerOpts: &slog.HandlerOptions{
				Level: slog.LevelWarn,
			},
		},
		TrafficLogger: &TrafficLogger{
			Output: "",
			LogFmt: LogFormatJSON,
		},
		HeaderFilters: NewHeaderFiltersContainer(),
		Cache:         cb,
	}
}
