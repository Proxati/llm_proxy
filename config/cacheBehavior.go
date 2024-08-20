package config

import (
	"log/slog"

	config_cache "github.com/proxati/llm_proxy/v2/config/cache"
)

type CacheEngine int

func (c CacheEngine) String() string {
	switch c {
	case CacheEngineMemory:
		return "memory"
	case CacheEngineBolt:
		return "bolt"
	default:
		return "memory"
	}
}

const (
	CacheEngineMemory CacheEngine = iota
	CacheEngineBolt
)

// cacheBehavior stores input args config for the cache
type cacheBehavior struct {
	Dir string // Directory to store the cache files
	// Size   int64  // Max size of the cache in total response records
	Engine      CacheEngine // Storage engine to use for cache
	EngineTitle string      // human-readable storage engine name (memory, bolt)
}

// NewCacheBehavior creates a new cacheBehavior object
func NewCacheBehavior(dir string, engineTitle string) *cacheBehavior {
	cb := &cacheBehavior{Dir: dir}
	cb.setEngine(engineTitle)
	return cb
}

// setEngine sets the cache engine enum based on the engineTitle string
func (c *cacheBehavior) setEngine(engineTitle string) {
	e := CacheEngineMemory
	switch engineTitle {
	case CacheEngineMemory.String():
		e = CacheEngineMemory
	case CacheEngineBolt.String():
		e = CacheEngineBolt
	}
	c.Engine = e
	c.EngineTitle = engineTitle
}

// GetCacheStorageConfig returns a cache.Config object based on the configured Engine
func (c *cacheBehavior) GetCacheStorageConfig(logger *slog.Logger) (config_cache.ConfigStorage, error) {
	switch c.Engine {
	case CacheEngineMemory:
		return config_cache.NewMemoryConfig(), nil
	case CacheEngineBolt:
		currentCacheConfigVer := "v1"
		cacheConfigFileName := "llm_proxy_cache.json"
		currentStorageVersion := "v1"
		defaultStorageEngineName := CacheEngineBolt.String()
		return config_cache.NewStorageJSON(
			logger,
			c.Dir,
			currentCacheConfigVer,
			cacheConfigFileName,
			currentStorageVersion,
			defaultStorageEngineName,
		)
	default:
		return config_cache.NewMemoryConfig(), nil
	}
}
