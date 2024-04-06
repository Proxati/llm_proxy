package cache

import (
	"fmt"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons/cache/storage/boltDB_Engine"
	"github.com/robbyt/llm_proxy/schema"
)

const (
	defaultBoltDBFile = "bolt.db"
)

// BoltMetaDB is a single boltDB with multiple internal "buckets" for each URL (like tables)
type BoltMetaDB struct {
	filterRespHeaders []string          // filter these headers when pulling from cache
	dbFileDir         string            // several DBs stored in the same directory, one for each base URL
	db                *boltDB_Engine.DB // the main db struct
	once              sync.Once
}

// len return the number of items currently in the cache
func (c *BoltMetaDB) len(identifier string) (int, error) {
	return c.db.Len(identifier)
}

// Close closes all the BadgerDBs in the collection
func (c *BoltMetaDB) Close() error {
	var err error
	c.once.Do(func() {
		err = c.db.Close()
	})
	return err
}

// Get receives a request, pulls out the request URL, uses that URL as a
// cache "identifier" (to use the correct storage DB), and then looks up the
// request in cache based on the body, returning the cached response if found.
//
// The request URL can be considered the primary index (different files per URL),
// and the body is the secondary index.
func (c *BoltMetaDB) Get(request *schema.TrafficObject) (response *schema.TrafficObject, err error) {
	if request.URL == nil || request.URL.String() == "" {
		return nil, fmt.Errorf("request URL is nil or empty")
	}

	identifier := request.URL.String()
	body := request.Body

	valueBytes, err := c.db.GetBytesSafe(identifier, []byte(body))
	if err != nil {
		return nil, err
	}
	if valueBytes == nil {
		log.Debugf("valueBytes empty for: %s", identifier)
		return nil, nil
	}

	newResponse, err := schema.NewFromJSONBytes(valueBytes, c.filterRespHeaders)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %s", err)
	}

	// return the cached response, as a traffic object
	return newResponse, nil
}

// Put receives a request and response, pulls out the request URL, uses that
// URL as a cache "identifier" (to use the correct storage DB), and then stores
// the response in cache based on the request body.
func (c *BoltMetaDB) Put(request *schema.TrafficObject, response *schema.TrafficObject) error {
	if request.URL == nil || request.URL.String() == "" {
		return fmt.Errorf("request URL is nil or empty")
	}
	identifier := request.URL.String()
	reqBody := []byte(request.Body)

	// Store the encoded data in the targetDB
	respJSON, err := response.ToJSON()
	if err != nil {
		return fmt.Errorf("error marshalling response object: %s", err)
	}

	err = c.db.SetBytes(identifier, reqBody, respJSON)
	if err != nil {
		log.Fatal("set bytes error:", err)
	}

	log.Debugf("stored response in cache for: %s", identifier)
	return nil
}

// NewBoltMetaDB creates a new BoltMetaDB object, to load or create a new boltDB on disk
func NewBoltMetaDB(dbFileDir string, filterRespHeaders []string) (*BoltMetaDB, error) {
	dbFile := filepath.Join(dbFileDir, defaultBoltDBFile)
	db, err := boltDB_Engine.NewDB(dbFile)
	if err != nil {
		return nil, fmt.Errorf("error opening/creating db: %s", err)
	}
	bMeta := &BoltMetaDB{
		filterRespHeaders: filterRespHeaders,
		dbFileDir:         dbFileDir,
		db:                db,
	}
	return bMeta, nil
}
