package cache

// MemoryConfig is a mostly empty struct that implements the Config interface for an in-memory cache
type MemoryConfig struct {
}

// NewMemoryConfig creates a new (empty) MemoryConfig object to implement the cache.ConfigStorage interface
func NewMemoryConfig() *MemoryConfig {
	return &MemoryConfig{}
}

// GetStorageEngine returns the storage engine type
func (m *MemoryConfig) GetStorageEngine() string {
	return "memory"
}

// GetStoragePath returns the storage path, which is RAM for this implementation
func (m *MemoryConfig) GetStoragePath() string {
	return "RAM"
}

// GetStorageVersion returns the storage version
func (m *MemoryConfig) GetStorageVersion() string {
	return "v1"
}

// Save writes the cache config, which is a noop for this implementation
func (m *MemoryConfig) Save() error {
	return nil
}

// Read loads the cache config, which is a noop for this implementation
func (m *MemoryConfig) Read() error {
	return nil
}
