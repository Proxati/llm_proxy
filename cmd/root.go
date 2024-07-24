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
var intro string = `
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

		// print the log splash screen
		if cfg.IsVerboseOrHigher() {
			if isatty.IsTerminal(os.Stdout.Fd()) {
				fmt.Print(intro)
			}
		}

		var err error
		cfg.TerminalSloggerFormat, err = config.StringToLogFormat(terminalLogFormat)
		cfg.SetLoggerLevel()
		slog.Debug("Global logger setup completed", "TerminalSloggerFormat", cfg.TerminalSloggerFormat.String())

		if err != nil {
			slog.Error("Could not setup terminal log", "error", err)
		}

		cfg.TrafficLogFmt, err = config.StringToLogFormat(trafficLogFormat)
		if err != nil {
			slog.Error("Could not setup traffic log", "error", err)
		}

		if err != nil {
			os.Exit(1)
		}
	},
	SilenceUsage: true,
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
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", cfg.Verbose, "Print runtime activity to stderr")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Debug, "debug", "d", cfg.Debug, "Print debug information to stderr")
	rootCmd.PersistentFlags().BoolVar(
		&cfg.Trace, "trace", cfg.Trace, "Print detailed trace debugging information to stderr, requires --debug to also be set")
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
		&cfg.OutputDir, "output", "o", "",
		"Directory to write request/response traffic logs (unset will write to stdout)",
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
