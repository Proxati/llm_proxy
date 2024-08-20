package cache

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _testFileName = "config.json"

func TestNewStorageJSON(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()
	tmpDir := t.TempDir()

	cacheConfig, err := NewStorageJSON(
		testLogger, tmpDir, "v1", _testFileName, "v1", "bolt")
	assert.NoError(t, err)
	assert.NotNil(t, cacheConfig)
	assert.Equal(t, tmpDir+"/cache", cacheConfig.StoragePath)

	// test loading from an existing file
	cacheConfig2, err := NewStorageJSON(
		testLogger, tmpDir, "v1", _testFileName, "v1", "bolt")
	assert.NoError(t, err)
	assert.NotNil(t, cacheConfig2)
	assert.Equal(t, cacheConfig.ConfigVersion, cacheConfig2.ConfigVersion)

	// update a value, and save it
	cacheConfig2.ConfigVersion = "42"
	err = cacheConfig2.Save()
	assert.NoError(t, err)

	// load the file again, and check the result from the loaded file
	cacheConfig3, err := NewStorageJSON(
		testLogger, tmpDir, "v1", _testFileName, "v1", "bolt")
	assert.NoError(t, err)
	assert.NotNil(t, cacheConfig3)
	assert.Equal(t, cacheConfig2.ConfigVersion, cacheConfig3.ConfigVersion)

	// test loading from a non-existing file
	newConfigName := tmpDir + "/non-existing"
	cacheConfig4, err := NewStorageJSON(
		testLogger, newConfigName, "v1", _testFileName, "v1", "bolt")
	assert.NoError(t, err)
	assert.NotNil(t, cacheConfig4)
	assert.Equal(t, newConfigName+"/"+_testFileName, cacheConfig4.filePath)

	// check the interface methods
	assert.Equal(t, newConfigName+"/cache", cacheConfig4.GetStoragePath())
	assert.Equal(t, "bolt", cacheConfig4.GetStorageEngine())
	assert.Equal(t, "v1", cacheConfig4.GetStorageVersion())

}

func TestStorageJSON_SaveAndLoad(t *testing.T) {
	t.Parallel()
	testLogger := slog.Default()
	tmpDir := t.TempDir()

	cacheConfig, _ := NewStorageJSON(
		testLogger, tmpDir, "v1", _testFileName, "v1", "bolt")
	err := cacheConfig.Save()
	assert.NoError(t, err)

	loadedCacheConfig := &JSONConfigFile{filePath: cacheConfig.filePath}
	err = loadedCacheConfig.Read()
	assert.NoError(t, err)

	assert.Equal(t, cacheConfig, loadedCacheConfig)
}
