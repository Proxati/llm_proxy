package boltDB_Engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/proxati/llm_proxy/v2/fileUtils"
	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/key"
	bolt "go.etcd.io/bbolt"
)

// DB is a wrapper for the DB database library
type DB struct {
	db        *bolt.DB
	closeOnce sync.Once
}

func (b *DB) Len(identifier string) (int, error) {
	count := 0
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(identifier))
		if bucket == nil {
			return BucketNotFoundError{Identifier: identifier}
		}

		bucket.ForEach(func(k, v []byte) error {
			count++
			return nil
		})

		return nil
	})
	return count, err
}

// GetBytes gets a value from the database using a byte key
func (b *DB) GetBytes(identifier string, key key.Key) (value []byte, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(identifier))
		if bucket == nil {
			return BucketNotFoundError{Identifier: identifier}
		}

		v := bucket.Get(key.Get())
		if v == nil {
			return ErrKeyNotFound{Identifier: identifier, Key: key.Get()}
		}

		// Make a copy of the value, so it's readable outside of this transaction
		value = make([]byte, len(v))
		copy(value, v)

		return nil
	})
	return value, err
}

// GetBytesSafe attempts to get a value from the database, and returns nil if not found
func (b *DB) GetBytesSafe(identifier string, key key.Key) ([]byte, error) {
	val, err := b.GetBytes(identifier, key)
	if err != nil {
		var keyNotFoundError ErrKeyNotFound
		var bucketNotFoundError BucketNotFoundError
		if errors.As(err, &keyNotFoundError) || errors.As(err, &bucketNotFoundError) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting value: %s", err)
	}
	return val, nil
}

// SetBytes sets a value in the database using a byte key
func (b *DB) SetBytes(identifier string, key key.Key, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(identifier))
		if err != nil {
			return fmt.Errorf("error creating/loading bucket: %s", err)
		}

		err = bucket.Put(key.Get(), value)
		if err != nil {
			return fmt.Errorf("error putting value: %s", err)
		}
		return nil
	})
}

// Close closes the database and runs other cleanup tasks
func (b *DB) Close() (err error) {
	b.closeOnce.Do(func() {
		errClose := b.db.Close()
		if errClose != nil {
			err = fmt.Errorf("error closing db: %s", errClose)
			return
		}
	})
	return err
}

func (b *DB) GetDBFileName() string {
	return b.db.Path()
}

func configBolt() *bolt.Options {
	return &bolt.Options{
		Timeout: 1 * time.Second,
	}
}

// NewDB creates a wrapper object for a NewDB database to creates new or load an existing DB.
// dbFileName: the path where the BoltDB file is stored on disk
func NewDB(dbFileName string) (*DB, error) {
	if dbFileName == "" {
		return nil, fmt.Errorf("db file name is empty")
	}

	dirPath := filepath.Dir(dbFileName)
	err := fileUtils.DirExistsOrCreate(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error creating db parent directory: %s", dirPath)
	}

	db, err := bolt.Open(dbFileName, 0600, configBolt())
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}
	return &DB{db: db}, nil
}
