package cmd

import (
	"github.com/spf13/cobra"

	"github.com/proxati/llm_proxy/config"
	"github.com/proxati/llm_proxy/proxy"
)

// simpleCmd represents a simple proxy server without logging
var simpleCmd = &cobra.Command{
	Use:   "simple",
	Short: "Run a simple LLM proxy server with no traffic caching.",
	Long:  "Simple LLM proxy. Use --verbose to log traffic to stdout. Use --output to save traffic logs to disk.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.SimpleMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(simpleCmd)
	simpleCmd.SuggestFor = simple_suggestions
}
