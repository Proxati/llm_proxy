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
	Long: `Services currently supported:
- OpenAI (Completions API only)

Disclaimer:
This tool is not affiliated with any of these APIs.
All information & calculations are an approximation, and should not be used for billing or budgeting purposes.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.APIAuditMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(apiAuditorCmd)
	apiAuditorCmd.SuggestFor = api_auditor_suggestions
}
