package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/storage"
	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/storage/boltDB_Engine"
)

func TestNewBadgerDB_CacheMap(t *testing.T) {
	cacheMap := storage.NewCacheMap()
	assert.NotNil(t, cacheMap)
	assert.Equal(t, 0, cacheMap.Len())
}

func TestBadgerDB_CacheMap_PutAndGet(t *testing.T) {
	cacheMap := storage.NewCacheMap()
	tempDir := t.TempDir()

	db, _ := boltDB_Engine.NewDB(tempDir + "/test.db")
	defer db.Close()

	cacheMap.Put("test", db)
	assert.Equal(t, 1, cacheMap.Len())

	retrievedDB, found := cacheMap.Get("test")
	assert.True(t, found)
	assert.Equal(t, db, retrievedDB)

	// access the entire map using All()
	all := cacheMap.All()
	assert.Equal(t, 1, len(all))
	assert.Equal(t, db, all["test"])
}

func TestBadgerDB_CacheMap_DeleteAndClear(t *testing.T) {
	cacheMap := storage.NewCacheMap()
	tempDir := t.TempDir()

	badgerDB, _ := boltDB_Engine.NewDB(tempDir + "/test.db")
	defer badgerDB.Close()

	cacheMap.Put("test", badgerDB)
	assert.Equal(t, 1, cacheMap.Len())

	cacheMap.Delete("test")
	assert.Equal(t, 0, cacheMap.Len())

	cacheMap.Put("test", badgerDB)
	assert.Equal(t, 1, cacheMap.Len())

	cacheMap.Clear()
	assert.Equal(t, 0, cacheMap.Len())
}
