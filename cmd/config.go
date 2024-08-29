package cmd

import "github.com/proxati/llm_proxy/v2/config"

// cfg is a reasonable default configuration, used by all commands
var cfg *config.Config = config.NewDefaultConfig()

// suggestions are here instead of their respective files bc it's easier to see them all in one place
var apiAuditorSuggestions = []string{
	"audit", "auditor", "api-auditor", "api-audit", "api-auditing",
}

var cacheSuggestions = []string{
	"cache-proxy", "caching-proxy", "cash-proxy", "cash",
}

var proxyRunSuggestions = []string{
	"proxy", "simple-proxy", "simpleproxy",
}
