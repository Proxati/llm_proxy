package boltDB_Engine

import (
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/key"
)

// Delete removes a key from the database
func (b *DB) Delete(identifier string, key key.Key) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(identifier))
		if err != nil {
			return fmt.Errorf("error creating/loading bucket: %s", err)
		}

		return bucket.Delete(key.Get())
	})
}
