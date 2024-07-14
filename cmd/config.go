package cmd

import "github.com/proxati/llm_proxy/config"

// cfg is a reasonable default configuration, used by all commands
var cfg *config.Config = config.NewDefaultConfig()

// suggestions are here instead of their respective files bc it's easier to see them all in one place
var api_auditor_suggestions = []string{
	"audit", "auditor", "api-auditor", "api-audit", "api-auditing",
}

var cache_suggestions = []string{
	"cache-proxy", "caching-proxy", "cash-proxy", "cash",
}

var proxyRun_suggestions = []string{
	"proxy", "simple-proxy", "simpleproxy",
}
