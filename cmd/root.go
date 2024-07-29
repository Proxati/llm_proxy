package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/proxati/llm_proxy/v2/config"
	"github.com/spf13/cobra"
)

// https://manytools.org/hacker-tools/ascii-banner/
const introSplashText = `
██████╗ ██████╗  ██████╗ ██╗  ██╗ █████╗ ████████╗██╗                                             
██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝██╔══██╗╚══██╔══╝██║                                             
██████╔╝██████╔╝██║   ██║ ╚███╔╝ ███████║   ██║   ██║                                             
██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗ ██╔══██║   ██║   ██║                                             
██║     ██║  ██║╚██████╔╝██╔╝ ██╗██║  ██║   ██║   ██║                                             
╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝                                             
                                                                                                  
██╗     ██╗     ███╗   ███╗        ██████╗ ██████╗  ██████╗ ██╗  ██╗██╗   ██╗    ██╗   ██╗██████╗ 
██║     ██║     ████╗ ████║        ██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝    ██║   ██║╚════██╗
██║     ██║     ██╔████╔██║        ██████╔╝██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝     ██║   ██║ █████╔╝
██║     ██║     ██║╚██╔╝██║        ██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝      ╚██╗ ██╔╝██╔═══╝ 
███████╗███████╗██║ ╚═╝ ██║███████╗██║     ██║  ██║╚██████╔╝██╔╝ ██╗   ██║        ╚████╔╝ ███████╗
╚══════╝╚══════╝╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝         ╚═══╝  ╚══════╝
                                                                                                  
`

// converted later to enum values in the config package
var terminalLogFormat string
var trafficLogFormat string
var debugMode bool
var verboseMode bool
var traceMode bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "llm_proxy",
	Short: "Proxy your LLM traffic for logging, security evaluation, and fine-tuning.",
	Long: `llm_proxy is an HTTP MITM (Man-In-The-Middle) proxy designed to log all requests and responses.

This is useful for:
  * Security: A multi-homed DMZ provides isolation between apps and external LLM APIs.
  * Debugging: Tag and observe all LLM API traffic.
  * Fine-tuning: Use the stored logs to fine-tune your LLM models.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupTerminalOutputLevel(cfg, debugMode, verboseMode, traceMode)

		logFormat, err := setupLogFormats(cfg, terminalLogFormat, trafficLogFormat)
		if err != nil {
			os.Exit(1)
		}

		s := printSplash(
			cfg.GetLoggerLevel(),
			cfg.GetTerminalOutputFormat(),
			isatty.IsTerminal(os.Stdout.Fd()),
			introSplashText,
		)
		if s != "" {
			fmt.Print(s)
		}
		cfg.GetLogger().Debug("Global logger setup completed", "TerminalSloggerFormat", logFormat.String())
	},
	SilenceUsage: true,
}

func setupTerminalOutputLevel(cfg *config.Config, debugMode, verboseMode, traceMode bool) {
	if debugMode {
		if traceMode {
			cfg.EnableOutputTrace()
		} else {
			cfg.EnableOutputDebug()
		}
	} else if verboseMode {
		cfg.EnableOutputVerbose()
	}
}

// printSplash will only show the logo if verbose mode is enabled, on a real terminal, with text output mode
func printSplash(logLevel slog.Level, logFormat config.LogFormat, isTTY bool, txt string) string {
	if !isTTY {
		return ""
	}

	if logFormat != config.LogFormat_TXT {
		return ""
	}

	switch logLevel {
	case slog.LevelDebug, slog.LevelInfo:
		return txt
	default:
		return ""
	}
}

// rootSetup always runs first, and configures the global logger and log formats
func setupLogFormats(cfg *config.Config, terminalLogFormat, trafficLogFormat string) (config.LogFormat, error) {

	// set the terminal log format, json or txt
	termLogFormat, termOutErr := cfg.SetTerminalOutputFormat(terminalLogFormat)
	if termOutErr != nil {
		cfg.SetTerminalOutputFormat("txt") // default to txt if there's an error
		slog.Error("Could not setup terminal log", "error", termOutErr)
	}

	// set the traffic log (to disk) format, json or txt
	trafficOutErr := cfg.SetTrafficLogFormat(trafficLogFormat)
	if trafficOutErr != nil {
		slog.Error("Could not setup traffic log", "error", trafficOutErr)
	}

	if termOutErr != nil || trafficOutErr != nil {
		return 0, fmt.Errorf("could not setup log formats")
	}

	return termLogFormat, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true // don't show the default completion command in help
	rootCmd.PersistentFlags().BoolVarP(
		&verboseMode, "verbose", "v", false, "Print runtime activity to stderr")
	rootCmd.PersistentFlags().BoolVarP(
		&debugMode, "debug", "d", false, "Print debug information to stderr")
	rootCmd.PersistentFlags().BoolVar(
		&traceMode, "trace", false, "Print detailed trace debugging information to stderr, requires --debug to also be set")
	rootCmd.PersistentFlags().MarkHidden("trace")

	rootCmd.PersistentFlags().StringVarP(
		&cfg.Listen, "listen", "l", cfg.Listen,
		"Address to listen on",
	)

	// Certificate Settings
	rootCmd.PersistentFlags().StringVarP(
		&cfg.CertDir, "ca_dir", "c", cfg.CertDir,
		"Path to the local trusted certificate, for TLS MITM",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&cfg.InsecureSkipVerifyTLS, "skip-upstream-tls-verify", "K", cfg.InsecureSkipVerifyTLS,
		"Skip upstream TLS cert verification",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&cfg.NoHttpUpgrader, "no-http-upgrader", "", cfg.NoHttpUpgrader,
		"Disable the automatic http->https request upgrader",
	)
	// Logging Settings
	rootCmd.PersistentFlags().StringVarP(
		&cfg.Output, "output", "o", "",
		`Comma-delimited list of log destinations. This can be a directory, or a
HTTP(s) REST API. If unset, and verbose/debug is enabled, traffic logs will be
sent to the terminal. See the documentation for more information.

Examples:
"/tmp/out", "file:///tmp/out", "http://my-api.com/log,/tmp/out"
`,
	)
	rootCmd.PersistentFlags().StringVar(
		&terminalLogFormat, "terminal-log-format", "txt",
		"Screen output format (valid options: json or txt)",
	)
	rootCmd.PersistentFlags().StringVar(
		&trafficLogFormat, "traffic-log-format", "json",
		"Disk output format for traffic logs (valid options: json or txt)",
	)
	rootCmd.PersistentFlags().BoolVar(
		&cfg.NoLogConnStats, "no-log-connection-stats", cfg.NoLogConnStats,
		"Don't write connection stats to traffic logs",
	)
	rootCmd.PersistentFlags().BoolVar(
		&cfg.NoLogReqHeaders, "no-log-req-headers", cfg.NoLogReqHeaders,
		"Don't write request headers to traffic logs",
	)
	rootCmd.PersistentFlags().BoolVar(
		&cfg.NoLogReqBody, "no-log-req-body", cfg.NoLogReqBody,
		"Don't write request body or details to traffic logs",
	)
	rootCmd.PersistentFlags().BoolVar(
		&cfg.NoLogRespHeaders, "no-log-resp-headers", cfg.NoLogRespHeaders,
		"Don't write response headers to traffic logs",
	)
	rootCmd.PersistentFlags().BoolVar(
		&cfg.NoLogRespBody, "no-log-resp-body", cfg.NoLogRespBody,
		"Don't write response body or details to traffic logs",
	)

	rootCmd.PersistentFlags().StringSliceVar(
		&cfg.FilterReqHeaders, "filter-req-headers", cfg.FilterReqHeaders,
		"Comma-separated list of request headers that will be omitted from logs",
	)
	rootCmd.PersistentFlags().MarkHidden("filter-req-headers")

	rootCmd.PersistentFlags().StringSliceVar(
		&cfg.FilterRespHeaders, "filter-resp-headers", cfg.FilterRespHeaders,
		"Comma-separated list of response headers that will be omitted from logs",
	)
	rootCmd.PersistentFlags().MarkHidden("filter-resp-headers")
}
