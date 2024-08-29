package cache

// ConfigStorage is the interface that all cache config objects must implement, it's used to save
// or load cache config, from disk, database, or just RAM for the in-memory cache.
type ConfigStorage interface {
	// GetStorageEngine returns the storage engine name used for this cache config
	GetStorageEngine() string

	// GetStoragePath returns the full path to the storage bucket (file path or database URI)
	GetStoragePath() string

	// GetStorageVersion returns the storage version used for this cache
	GetStorageVersion() string

	// Save writes the cache config
	Save() error

	// Read loads the cache config
	Read() error
}
