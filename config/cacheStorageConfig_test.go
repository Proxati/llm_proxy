package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheStorageConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cacheConfig, err := NewCacheStorageConfig(tmpDir)
	assert.NoError(t, err)
	assert.NotNil(t, cacheConfig)
	assert.Equal(t, tmpDir+"/cache", cacheConfig.StoragePath)

	// test loading from an existing file
	cacheConfig2, err := NewCacheStorageConfig(tmpDir)
	assert.NoError(t, err)
	assert.NotNil(t, cacheConfig2)
	assert.Equal(t, cacheConfig.ConfigVersion, cacheConfig2.ConfigVersion)

	// update a value, and save it
	cacheConfig2.ConfigVersion = "42"
	err = cacheConfig2.Save()
	assert.NoError(t, err)

	// load the file again, and check the result from the loaded file
	cacheConfig3, err := NewCacheStorageConfig(tmpDir)
	assert.NoError(t, err)
	assert.NotNil(t, cacheConfig3)
	assert.Equal(t, cacheConfig2.ConfigVersion, cacheConfig3.ConfigVersion)

}

func TestCacheStorageConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	cacheConfig, _ := NewCacheStorageConfig(tmpDir)
	err := cacheConfig.Save()
	assert.NoError(t, err)

	loadedCacheConfig := &CacheStorageConfig{filePath: cacheConfig.filePath}
	err = loadedCacheConfig.Load()
	assert.NoError(t, err)

	assert.Equal(t, cacheConfig, loadedCacheConfig)
}
