package cmd

import (
	"fmt"
	"os"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/spf13/cobra"
)

// string variable that will be converted to an enum in the config package
var logFormatStr = "json"

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
		// setup logger
		cfg.SetLoggerLevel()

		// setup the traffic log format, load string to enum
		var err error
		cfg.LogFormat, err = config.StringToTrafficLogFormat(logFormatStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not setup traffic log: %v\n", err)
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
		&logFormatStr, "traffic-log-format", "json",
		"Traffic log output format (json, txt)",
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
