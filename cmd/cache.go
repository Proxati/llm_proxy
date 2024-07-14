package cmd

import (
	"github.com/spf13/cobra"

	"github.com/proxati/llm_proxy/config"
	"github.com/proxati/llm_proxy/proxy"
)

// cacheCmd represents the mock command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Enable literal caching mode to store all traffic in a local embedded database.",
	Long: `This command creates a proxy server that inspects each request body and responds
from the cache if an identical request body is found. When no cached copy is available, the
server forwards the request to the upstream server and stores the response in a local
directory for future use. This caching mechanism is ideal for development and CI, reducing
the number of requests to the upstream server. The cache stores responses with the same status
code, headers, and body as the original, except for responses with status codes 500 or higher.`,
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
