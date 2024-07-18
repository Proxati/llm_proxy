package memory_Engine

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/key"
)

// MemoryStorage is a simple in-memory storage engine
type MemoryStorage struct {
	cache *lru.Cache[string, []byte]
}

// GetBytes gets a value from the database using a byte key
func (m *MemoryStorage) GetBytes(identifier string, key key.Key) ([]byte, error) {
	val, ok := m.cache.Get(key.String())
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key.String())
	}

	return val, nil
}

// GetBytesSafe attempts to get a value from the database, and returns nil if not found
func (m *MemoryStorage) GetBytesSafe(identifier string, key key.Key) ([]byte, error) {
	val, ok := m.cache.Get(key.String())
	if !ok {
		return nil, nil
	}

	return val, nil
}

// SetBytes sets a value in the database using a byte key
func (m *MemoryStorage) SetBytes(identifier string, key key.Key, value []byte) error {
	m.cache.Add(key.String(), value)
	return nil
}

// Close closes the database
func (m *MemoryStorage) Close() error {
	m.cache = nil
	return nil
}

// GetDBFileName returns the on-disk filename of the database
func (m *MemoryStorage) GetDBFileName() string {
	return "RAM"
}

// NewMemoryStorage creates a new MemoryStorage object
func NewMemoryStorage(maxEntries int) (*MemoryStorage, error) {
	cache, err := lru.New[string, []byte](maxEntries)
	if err != nil {
		return nil, err
	}
	return &MemoryStorage{cache: cache}, nil
}
