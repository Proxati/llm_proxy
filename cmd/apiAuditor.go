package cmd

import (
	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy"
	"github.com/spf13/cobra"
)

// apiAuditorCmd represents the apiAuditor command
var apiAuditorCmd = &cobra.Command{
	Use:   "apiAuditor",
	Short: "A realtime view of how much you are spending on 3rd party AI services",
	Long: `Provides a real-time view of your spending on third-party AI services.

## Services Currently Supported
- OpenAI (Completions API only)

## Important Disclaimer
This tool is not affiliated with any of these APIs. All billing information is an approximation
based on the latest available information. The calculations made are approximations and should
not be used for billing or budgeting purposes.

## Features
- Real-time Cost Monitoring: Track the cost of API calls in real-time.

## Example Usage

# Start the apiAuditor with default settings
./llm_proxy apiAuditor

# Start the apiAuditor with verbose logging
./llm_proxy apiAuditor --verbose
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.APIAuditMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(apiAuditorCmd)
	apiAuditorCmd.SuggestFor = apiAuditorSuggestions
}
