package cmd

import (
	"github.com/spf13/cobra"

	"github.com/proxati/llm_proxy/config"
	"github.com/proxati/llm_proxy/proxy"
)

// cacheCmd represents the mock command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Enable caching mode, which stores all requests / responses in a local directory",
	Long: `This command creates a proxy server that sends responses to the upstream server only
when there isn't a copy available in the cache. The cache command requires a local directory to store
and retrieve the responses. This mode is useful for development and for CI, because it will reduce the
number of requests to the upstream server. The cache server will respond with the same status code,
headers, and body as the previous response. The cache server will not store responses with a status
code of 500 or higher.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.CacheMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.SuggestFor = cache_suggestions

	cacheCmd.Flags().StringVar(
		&cfg.Cache.Dir, "cache-dir", cfg.Cache.Dir,
		"Directory to store the cache database files",
	)
	/*
		cacheCmd.Flags().Int64VarP(
			&cfg.Cache.TTL, "ttl", "", cfg.Cache.TTL,
			"Time to live for cache files in seconds (0 means cache forever)",
		)
	*/
	/*
		cacheCmd.Flags().Int64Var(
			&cfg.Cache.TTL, "max", cfg.Cache.MaxRecords,
			"Limit # of cached records. LRU deletion. 0=no limit. ",
		)
	*/
}
