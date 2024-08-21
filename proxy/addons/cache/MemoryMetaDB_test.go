package cache

import (
	"encoding/json"
	"net/url"
	"testing"

	"log/slog"

	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryMetaDB_NewMemoryMetaDB(t *testing.T) {
	logger := slog.Default()
	db, err := NewMemoryMetaDB(logger, 10)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
}

func TestMemoryMetaDB_Close(t *testing.T) {
	logger := slog.Default()
	db, err := NewMemoryMetaDB(logger, 10)
	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err)
}

func TestMemoryMetaDB_Len(t *testing.T) {
	logger := slog.Default()
	db, err := NewMemoryMetaDB(logger, 10)
	require.NoError(t, err)
	defer db.Close()

	requestURL, err := url.Parse("http://example.com")
	require.NoError(t, err)

	request := &schema.ProxyRequest{
		URL:  requestURL,
		Body: "request body",
	}

	response := &schema.ProxyResponse{
		Status: 200,
		Body:   "response body",
	}

	err = db.Put(request, response)
	require.NoError(t, err)

	length, err := db.Len(request.URL.String())
	require.NoError(t, err)
	assert.Equal(t, 1, length)
}

func TestMemoryMetaDB_Get(t *testing.T) {
	logger := slog.Default()
	db, err := NewMemoryMetaDB(logger, 10)
	require.NoError(t, err)
	defer db.Close()

	requestURL, err := url.Parse("http://example.com")
	require.NoError(t, err)

	request := &schema.ProxyRequest{
		URL:  requestURL,
		Body: "request body", // Use string instead of []byte
	}

	response := &schema.ProxyResponse{
		Status: 200,             // Correct field name
		Body:   "response body", // Use string instead of []byte
		Header: nil,             // Ensure header is nil
	}

	err = db.Put(request, response)
	require.NoError(t, err)

	storedResponse, err := db.Get(request.URL.String(), []byte(request.Body)) // Convert request.Body to []byte
	require.NoError(t, err)
	require.NotNil(t, storedResponse)

	// Ensure the header is nil if it is an empty map
	if len(storedResponse.Header) == 0 {
		storedResponse.Header = nil
	}

	expectedResponseJSON, err := json.Marshal(response)
	require.NoError(t, err)

	storedResponseJSON, err := json.Marshal(storedResponse)
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedResponseJSON), string(storedResponseJSON))
}

func TestMemoryMetaDB_GetOrCreateDb(t *testing.T) {
	logger := slog.Default()
	db, err := NewMemoryMetaDB(logger, 10)
	require.NoError(t, err)
	defer db.Close()

	storage, err := db.getOrCreateDb("test_identifier")
	require.NoError(t, err)
	require.NotNil(t, storage)
}
