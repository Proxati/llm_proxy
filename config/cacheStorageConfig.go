package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/proxati/llm_proxy/v2/fileUtils"
)

const (
	currentCacheConfigVer    = "v1"
	cacheConfigFileName      = "llm_proxy_cache.json"
	currentStorageVersion    = "v1"
	defaultStorageEngineName = "bolt"
)

// CacheStorage is a struct that backs a llm_proxy_cache.json file, which configures the cache storage object
type CacheStorage struct {
	filePath       string `json:"-"`               // The full path of this cache index json file.
	ConfigVersion  string `json:"config_version"`  // The schema version of this cache index file.
	StorageEngine  string `json:"storage_engine"`  // The storage engine used for this cache
	StorageVersion string `json:"storage_version"` // The storage version used for this cache
	StoragePath    string `json:"storage_path"`    // The full path to the storage bucket (file path or database URI)
}

// Save writes the cache config json file to disk
func (i CacheStorage) Save() error {
	// Ensure the storage path subdirectory exists
	if err := os.MkdirAll(filepath.Dir(i.StoragePath), 0700); err != nil {
		return err
	}

	// Set the schema version if it's not already set
	if i.ConfigVersion == "" {
		i.ConfigVersion = currentCacheConfigVer
	}

	// Convert the IndexFile object to a JSON string
	jsonData, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON string to a tmp file, then rename it to the final file path
	tmpFile, err := os.CreateTemp(filepath.Dir(i.filePath), cacheConfigFileName)
	if err != nil {
		return err
	}

	// Defer closing and deleting the tmp file, in case of an error
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if err = os.WriteFile(tmpFile.Name(), jsonData, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile.Name(), i.filePath)
}

// Load reads the cache config json file from disk
func (i *CacheStorage) Load() error {
	existingFilePath := i.filePath
	jsonData, err := os.ReadFile(existingFilePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonData, i); err != nil {
		return err
	}

	i.filePath = existingFilePath
	return nil
}

// NewCacheStorageConfig creates a new IndexFile object to help with loading/saving meta-state as a json file.
// This object's purpose is to help loading the other database objects by pointing to their
// connection settings or file paths.
//
// cacheDir: the directory where the cache index file will be stored
func NewCacheStorageConfig(logger *slog.Logger, cacheDir string) (*CacheStorage, error) {
	// this is a special case where we're NOT using a package-level slog.go, because the config
	// package also configures the logger.
	logger = logger.WithGroup("config.CacheStorage")

	indexFilePath := filepath.Join(cacheDir, cacheConfigFileName)
	iFile := &CacheStorage{
		filePath:       indexFilePath,
		ConfigVersion:  currentCacheConfigVer,
		StorageEngine:  defaultStorageEngineName,
		StorageVersion: currentStorageVersion,
		StoragePath:    filepath.Join(cacheDir, "cache"),
	}

	if fileUtils.FileExists(iFile.filePath) {
		logger.Debug("Loading existing cache config file", "filePath", iFile.filePath)
		if err := iFile.Load(); err != nil {
			return nil, fmt.Errorf("failed to load cache config file: %s", err)
		}
		logger.Debug("Loaded cache config file", "iFile", iFile)
		return iFile, nil
	}

	err := iFile.Save()
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %s", err)
	}
	logger.Info("Created a new cache config file", "filePath", iFile.filePath)
	return iFile, nil
}
