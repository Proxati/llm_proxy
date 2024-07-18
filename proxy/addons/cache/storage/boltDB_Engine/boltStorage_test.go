package boltDB_Engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/key"
)

func TestNewBoltDB(t *testing.T) {
	t.Run("valid db file", func(t *testing.T) {
		tempDir := t.TempDir()
		testDB := tempDir + "/test.db"

		db, err := NewDB(testDB)
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, testDB, db.GetDBFileName())
	})

	t.Run("invalid db file", func(t *testing.T) {
		_, err := NewDB("")
		assert.Error(t, err)
	})

	/*
		t.Run("close error from deleted db file", func(t *testing.T) {
			tempDir := t.TempDir()
			testDB := tempDir + "/test.db"

			db, err := NewBoltDB(testDB)
			require.NoError(t, err)

			// delete the file, and close should return an error
			// TODO: not failing as expected
			os.RemoveAll(testDB)
			err = db.Close()
			assert.Error(t, err)
		})
	*/
}

func TestBoltDB_GetSetStr(t *testing.T) {
	tempDir := t.TempDir()
	testDB := tempDir + "/test.db"

	db, err := NewDB(testDB)
	require.NoError(t, err)
	testKey := key.NewKeyStr("key")
	emptyKey := key.NewKeyStr("")
	defer db.Close()

	t.Run("normal set and get", func(t *testing.T) {
		err = db.SetBytes(t.Name(), testKey, []byte("value"))
		require.NoError(t, err)

		val, err := db.GetBytes(t.Name(), testKey)
		require.NoError(t, err)
		assert.Equal(t, "value", string(val))

		val, err = db.GetBytesSafe(t.Name(), testKey)
		require.NoError(t, err)
		assert.Equal(t, "value", string(val))

		err = db.Delete(t.Name(), testKey)
		require.NoError(t, err)

		val, err = db.GetBytesSafe(t.Name(), testKey)
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("set and get empty string", func(t *testing.T) {
		err = db.SetBytes(t.Name(), testKey, []byte(""))
		require.NoError(t, err)

		val, err := db.GetBytes(t.Name(), testKey)
		require.NoError(t, err)
		assert.Equal(t, "", string(val))
	})

	t.Run("set and get empty key", func(t *testing.T) {
		err = db.SetBytes(t.Name(), emptyKey, []byte("value"))
		require.NoError(t, err)

		val, err := db.GetBytes(t.Name(), emptyKey)
		require.NoError(t, err)
		assert.Equal(t, "value", string(val))
	})

	t.Run("set and get empty key and value", func(t *testing.T) {
		err = db.SetBytes(t.Name(), emptyKey, []byte(""))
		require.NoError(t, err)

		val, err := db.GetBytes(t.Name(), emptyKey)
		require.NoError(t, err)
		assert.Equal(t, "", string(val))
	})

	t.Run("set and get empty bucket", func(t *testing.T) {
		err = db.SetBytes("", emptyKey, []byte("value"))
		require.Error(t, err)
	})

	t.Run("set and get empty bucket and key", func(t *testing.T) {
		err = db.SetBytes("", emptyKey, []byte("value"))
		require.Error(t, err)
	})

	t.Run("set and get empty bucket and key and value", func(t *testing.T) {
		err = db.SetBytes("", emptyKey, []byte(""))
		require.Error(t, err)
	})

	t.Run("set and get using StrSafe", func(t *testing.T) {
		val, err := db.GetBytesSafe(t.Name(), emptyKey)
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	/*
		// expected to fail, but not failing after deleting the db file
		t.Run("set and get with deleted database using StrSafe", func(t *testing.T) {
			tempDir := t.TempDir()
			testDB := tempDir + "/test.db"

			db, err := NewBoltDB(testDB)
			require.NoError(t, err)

			err = db.SetStr(t.Name(), "key", "value")
			require.NoError(t, err)

			os.RemoveAll(testDB)
			os.RemoveAll(tempDir)

			val, err := db.GetStrSafe(t.Name(), "key")
			require.Error(t, err)
			assert.Nil(t, val)
		})
	*/
}

func TestBoltDB_GetSetBytes(t *testing.T) {
	tempDir := t.TempDir()
	testDB := tempDir + "/test.db"
	testKey := key.NewKeyStr("key")

	db, err := NewDB(testDB)
	require.NoError(t, err)
	defer db.Close()

	err = db.SetBytes("test", testKey, []byte("value"))
	require.NoError(t, err)

	val, err := db.GetBytes("test", testKey)
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)
}
