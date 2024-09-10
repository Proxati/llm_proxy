package cache

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/key"
	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/storage/memory_Engine"
	"github.com/proxati/llm_proxy/v2/schema"
)

type MemoryMetaDB struct {
	metaDB          map[string]*memory_Engine.MemoryStorage
	maxEntriesPerID int
	mutex           sync.RWMutex
	logger          *slog.Logger
}

// NewMemoryMetaDB creates a new MemoryMetaDB object
func NewMemoryMetaDB(logger *slog.Logger, maxEntries int) (*MemoryMetaDB, error) {
	return &MemoryMetaDB{
		metaDB:          make(map[string]*memory_Engine.MemoryStorage),
		maxEntriesPerID: maxEntries,
		logger:          logger.WithGroup("MemoryMetaDB"),
	}, nil
}

// String returns a string representation of the MemoryMetaDB object
func (c *MemoryMetaDB) String() string {
	return fmt.Sprintf("MemoryMetaDB: %d entries", len(c.metaDB))
}

func (c *MemoryMetaDB) Close() error {
	for _, db := range c.metaDB {
		db.Close()
	}
	return nil
}

func (c *MemoryMetaDB) Len(identifier string) (int, error) {
	if db, ok := c.metaDB[identifier]; ok {
		return db.Len(), nil
	}
	return 0, nil
}

func (c *MemoryMetaDB) Get(identifier string, body []byte) (response *schema.ProxyResponse, err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	db, ok := c.metaDB[identifier]
	if !ok {
		// no cache for this identifier (never seen this url base previously)
		return nil, nil
	}
	valueBytes, err := db.GetBytesSafe(identifier, key.NewKey(body))
	if err != nil {
		return nil, fmt.Errorf("could not read bytes from memory: %w", err)
	}
	if valueBytes == nil {
		c.logger.Debug("valueBytes empty", "identifier", identifier)
		return nil, nil
	}
	newResponse, err := schema.NewProxyResponseFromJSONBytes(valueBytes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %s", err)
	}

	// return the cached response, as a traffic object
	return newResponse, nil
}

// getOrCreateDb returns the memory storage for the given identifier, creating it if it doesn't exist
func (c *MemoryMetaDB) getOrCreateDb(identifier string) (*memory_Engine.MemoryStorage, error) {
	c.mutex.RLock()
	db, ok := c.metaDB[identifier]
	c.mutex.RUnlock()
	if ok {
		// found an existing db for this identifier
		return db, nil
	}

	// upgrade to a write lock
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// recheck if the identifier was created by another goroutine
	db, ok = c.metaDB[identifier]
	if ok {
		// found it, avoided the race condition!
		return db, nil
	}

	// nothing in the metadb for this identifier, create it
	db, err := memory_Engine.NewMemoryStorage(c.logger, identifier, c.maxEntriesPerID)
	if err != nil {
		return nil, fmt.Errorf("could not create memory storage: %w", err)
	}
	c.metaDB[identifier] = db
	return db, nil
}

func (c *MemoryMetaDB) Put(request *schema.ProxyRequest, response *schema.ProxyResponse) error {
	if request.URL == nil || request.URL.String() == "" {
		return fmt.Errorf("request URL is nil or empty")
	}
	identifier := request.URL.String()
	c.logger.Debug("Put", "identifier", identifier)

	db, err := c.getOrCreateDb(identifier)
	if err != nil {
		return fmt.Errorf("could not get or create db: %w", err)
	}
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	// store the response in the cache
	respJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling response object: %s", err)
	}
	/*
		slog.Default().Debug(
			"storing response in cache",
			"identifier", identifier,
			"response", string(respJSON),
			"key", key.NewKeyStr(request.Body).String(),
		)
	*/

	if err := db.SetBytes(identifier, key.NewKeyStr(request.Body), respJSON); err != nil {
		return fmt.Errorf("could not store response in cache: %w", err)
	}
	return nil
}
