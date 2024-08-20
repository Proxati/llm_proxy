package config

import (
	"fmt"
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
		return ""
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
	Engine CacheEngine // Storage engine to use for cache
}

// NewCacheBehavior creates a new cacheBehavior object
func NewCacheBehavior(dir string, engineTitle string) (*cacheBehavior, error) {
	cb := &cacheBehavior{Dir: dir}
	err := cb.SetEngine(engineTitle)
	if err != nil {
		return nil, fmt.Errorf("unable to create new cache behavior object: %w", err)
	}

	return cb, nil
}

// SetEngine sets the cache engine enum based on the engineTitle string
func (c *cacheBehavior) SetEngine(engineTitle string) error {
	switch engineTitle {
	case CacheEngineMemory.String():
		c.Engine = CacheEngineMemory
	case CacheEngineBolt.String():
		c.Engine = CacheEngineBolt
	default:
		return fmt.Errorf("invalid cache engine: %s", engineTitle)
	}
	return nil
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
		return nil, fmt.Errorf("invalid cache engine: %s", c.Engine)
	}
}
