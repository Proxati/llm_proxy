package cmd

import (
	"github.com/spf13/cobra"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy"
)

var cacheEngineTitle string = "bolt"

// cacheCmd represents the mock command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Enable literal caching mode to store all traffic in a local embedded database.",
	Long: `This command creates a proxy server that inspects each request body and responds
from the cache if an identical request body is found. When no cached copy is available, the
server forwards the request to the upstream server and stores the response in a local
directory for future use. This caching mechanism is ideal for development and CI, reducing
the number of requests to the upstream server.

## Features
- Storage Engines: Supports multiple storage engines, including in-memory and
BoltDB. The storage engine can be configured using the '--cache-engine' flag.
- Cache-Control Headers: Honors the 'Cache-Control' headers in the request and
response. If the request has a 'Cache-Control=no-cache' header, this proxy will
bypass the cache and forward the request to the upstream server.
- Portable Cache Directory: The cache directory can be moved between CPU
architectures and operating systems, because the storage engine is written in 100%
Golang. The default cache directory is "/tmp/llm_proxy", so it is recommended to
manually set this directory to a better location using the '--cache-dir' flag.

## Example Usage

# Start the proxy server with in-memory caching
./llm_proxy cache --cache-engine memory

# Start the proxy server with BoltDB caching
./llm_proxy cache --cache-engine bolt --cache-dir /var/cache/llm_proxy
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.CacheMode
		err := cfg.Cache.SetEngine(cacheEngineTitle)
		if err != nil {
			return err
		}
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.SuggestFor = cacheSuggestions

	cacheCmd.Flags().StringVar(
		&cfg.Cache.Dir, "cache-dir", cfg.Cache.Dir,
		"Directory to store the cache database files",
	)
	cacheCmd.Flags().StringVar(
		&cacheEngineTitle, "cache-engine", cacheEngineTitle,
		`Storage engine to use for cache (memory, bolt). When using bolt, the
cache-dir must be a valid writable path.`,
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
