package config

import (
	"testing"

	"log/slog"

	cache_config "github.com/proxati/llm_proxy/v2/config/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCacheStorageConfig(t *testing.T) {
	t.Parallel()
	logger := slog.Default()

	t.Run("MemoryEngine", func(t *testing.T) {
		cb, err := newCacheBehavior("/tmp", "memory")
		require.NotNil(t, cb)
		require.NoError(t, err)

		config, err := cb.GetCacheStorageConfig(logger)
		require.NoError(t, err)
		assert.IsType(t, &cache_config.MemoryConfig{}, config)
	})

	t.Run("BoltEngine", func(t *testing.T) {
		cb, err := newCacheBehavior("/tmp", "bolt")
		require.NotNil(t, cb)
		require.NoError(t, err)

		config, err := cb.GetCacheStorageConfig(logger)
		require.NoError(t, err)
		assert.IsType(t, &cache_config.JSONConfigFile{}, config)
	})

	t.Run("InvalidEngine", func(t *testing.T) {
		cb, err := newCacheBehavior("/tmp", "invalid")
		require.Error(t, err)
		require.Nil(t, cb)
	})
}
