package memory_Engine

import (
	"fmt"
	"log/slog"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/key"
)

// MemoryStorage is a simple in-memory storage engine
type MemoryStorage struct {
	name   string
	cache  *lru.TwoQueueCache[string, []byte]
	logger *slog.Logger
}

// NewMemoryStorage creates a new MemoryStorage object
func NewMemoryStorage(logger *slog.Logger, name string, maxEntries int) (*MemoryStorage, error) {
	cache, err := lru.New2Q[string, []byte](maxEntries)
	if err != nil {
		return nil, err
	}
	return &MemoryStorage{
		name:   name,
		cache:  cache,
		logger: logger.WithGroup("MemoryStorage").With("name", name),
	}, nil
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
func (m *MemoryStorage) SetBytes(_ string, key key.Key, value []byte) error {
	ks := key.String()
	m.cache.Add(ks, value)
	m.logger.Debug("set", "key", ks, "value", string(value))
	return nil
}

// Close closes the database
func (m *MemoryStorage) Close() error {
	m.cache = nil
	m.logger.Debug("Closed")
	return nil
}

// GetDBFileName returns the on-disk filename of the database
func (m *MemoryStorage) GetDBFileName() string {
	return "RAM"
}

// Len returns the number of items currently in the cache
func (m *MemoryStorage) Len() int {
	return m.cache.Len()
}
