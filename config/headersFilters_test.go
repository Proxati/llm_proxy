package config

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderFilterGroup(t *testing.T) {
	t.Parallel()
	userHeaders := []string{"User-Agent", "Accept"}
	extraHeaders1 := []string{"Foo"}
	extraHeaders2 := []string{"Bar"}
	hfg := NewHeaderFilterGroup(t.Name(), userHeaders, extraHeaders1, extraHeaders2)

	// Check initialization
	assert.Equal(
		t, append(extraHeaders1, extraHeaders2...),
		hfg.additionalHeaders,
		"additionalHeaders should be initialized correctly",
	)
	assert.Equal(t, userHeaders, hfg.Headers, "userHeaders should be initialized correctly")
	assert.NotNil(t, hfg.index, "Index should be initialized")

	// Check IsHeaderInGroup method
	assert.True(t, hfg.IsHeaderInGroup("Foo"), "Foo header should be in the index")
	assert.True(t, hfg.IsHeaderInGroup("Bar"), "Bar header should be in the index")
	assert.True(t, hfg.IsHeaderInGroup("User-Agent"), "User-Agent header should be in the index")
	assert.True(t, hfg.IsHeaderInGroup("Accept"), "Accept header should be in the index")
	assert.False(t, hfg.IsHeaderInGroup("Non-Existent-Header"), "Non-Existent-Header should not be in the index")

	// Check FilterHeaders method with variadic arguments
	reqHeaders := http.Header{
		"Authorization": {"Bearer token"},
		"Cookie":        {"session_id=123"},
		"User-Agent":    {"Go-http-client/1.1"},
		"Accept":        {"*/*"},
		"Foo":           {"value"},
		"Bar":           {"value"},
		"Faz":           {"Baz"},
	}

	// Test filtering with additional headers
	filteredHeaders := hfg.FilterHeaders(reqHeaders, "Authorization", "Cookie")
	assert.NotContains(t, filteredHeaders, "Authorization", "filtered from the variadic arguments")
	assert.NotContains(t, filteredHeaders, "Cookie", "filtered from the variadic arguments")
	assert.NotContains(t, filteredHeaders, "User-Agent", "filtered from the userHeaders")
	assert.NotContains(t, filteredHeaders, "Accept", "filtered from the userHeaders")
	assert.NotContains(t, filteredHeaders, "Foo", "filtered from the additionalHeaders")
	assert.NotContains(t, filteredHeaders, "Bar", "filtered from the additionalHeaders")
	assert.Contains(t, filteredHeaders, "Faz", "Faz header should not be filtered out")

	// Test filtering without additional headers
	filteredHeaders = hfg.FilterHeaders(reqHeaders)
	assert.Contains(t, filteredHeaders, "Authorization", "Authorization should not be filtered out")
	assert.Contains(t, filteredHeaders, "Cookie", "Cookie should not be filtered out")
	assert.NotContains(t, filteredHeaders, "User-Agent", "filtered from the userHeaders")
	assert.NotContains(t, filteredHeaders, "Accept", "filtered from the userHeaders")
	assert.NotContains(t, filteredHeaders, "Foo", "filtered from the additionalHeaders")
	assert.NotContains(t, filteredHeaders, "Bar", "filtered from the additionalHeaders")
	assert.Contains(t, filteredHeaders, "Faz", "Faz header should not be filtered out")
}

func TestNewHeaderFiltersContainer(t *testing.T) {
	t.Parallel()
	hfc := NewHeaderFiltersContainer()

	assert.NotNil(t, hfc.RequestToLogs, "RequestToLogs should be initialized")
	assert.NotNil(t, hfc.ResponseToLogs, "ResponseToLogs should be initialized")
	assert.NotNil(t, hfc.RequestToUpstream, "RequestToUpstream should be initialized")
	assert.NotNil(t, hfc.ResponseToClient, "ResponseToClient should be initialized")

	// index should be initialized after creation of the container
	assert.True(t, hfc.RequestToLogs.IsHeaderInGroup("Authorization"), "Authorization header should be in the RequestToLogs index")
	assert.True(t, hfc.ResponseToLogs.IsHeaderInGroup("Set-Cookie"), "Set-Cookie header should be in the ResponseToLogs index")

	hfc.RequestToLogs.Headers = []string{}
	hfc.ResponseToLogs.Headers = []string{}

	assert.True(t, hfc.RequestToLogs.IsHeaderInGroup("Authorization"), "Stale index")
	assert.True(t, hfc.ResponseToLogs.IsHeaderInGroup("Set-Cookie"), "Stale index")
	hfc.BuildIndexes()
	assert.False(t, hfc.RequestToLogs.IsHeaderInGroup("Authorization"), "Authorization header should not be in the RequestToLogs index")
	assert.False(t, hfc.ResponseToLogs.IsHeaderInGroup("Set-Cookie"), "Set-Cookie header should not be in the ResponseToLogs index")
}

func TestEmptyFull(t *testing.T) {
	t.Parallel()
	hfc := NewHeaderFiltersContainer()

	assert.NotNil(t, hfc.RequestToLogs.index, "RequestToLogs index should be initialized")
	assert.NotEmpty(t, hfc.RequestToLogs.index, "RequestToLogs index should have values")

	assert.NotNil(t, hfc.ResponseToLogs.index, "ResponseToLogs index should be initialized")
	assert.NotEmpty(t, hfc.ResponseToLogs.index, "ResponseToLogs index should have values")

	assert.NotNil(t, hfc.RequestToUpstream.index, "RequestToUpstream index should be initialized")
	assert.Empty(t, hfc.RequestToUpstream.index, "RequestToUpstream index should be be empty")

	assert.NotNil(t, hfc.ResponseToClient.index, "ResponseToClient index should be initialized")
	assert.Empty(t, hfc.ResponseToClient.index, "ResponseToClient index should be empty")
}
