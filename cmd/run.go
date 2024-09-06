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
	Long: `Starts llm_proxy in normal Man-in-the-Middle (MiTM) mode. This mode captures all
traffic passing through the proxy, which can be either printed to stdout or saved to disk.

## Features
- MiTM Mode: Intercepts and captures all HTTP/HTTPS traffic passing through the proxy.
- Traffic Logging: Optionally log captured traffic to stdout or save it to disk.
- Configurable Logging Levels: Adjust the verbosity of logs to suit your needs.

## Common Configuration Options
- --verbose: Show captured traffic on stdout.
- --output: Specify a file to save traffic logs.

## Example Usage

# Start the proxy server and print captured traffic to stdout
./llm_proxy run --verbose

# Start the proxy server and save captured traffic to a specific directory
./llm_proxy run --output /tmp/logs
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.ProxyRunMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(proxyRunCmd)
	proxyRunCmd.SuggestFor = proxyRunSuggestions
}
