package cmd

import (
	"github.com/spf13/cobra"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy"
)

// proxyRunCmd represents a simple proxy server without logging
var proxyRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the LLM proxy server with no traffic caching, and optional traffic logging.",
	Long:  "Use --verbose to show traffic on stdout. Use --output to save traffic logs to disk.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.ProxyRunMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(proxyRunCmd)
	proxyRunCmd.SuggestFor = proxyRun_suggestions
}
