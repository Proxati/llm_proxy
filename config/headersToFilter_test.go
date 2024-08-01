package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHeaderFilterGroup(t *testing.T) {
	t.Parallel()
	headers := []string{"Authorization", "Cookie"}
	hfg := NewHeaderFilterGroup(headers)

	assert.Equal(t, headers, hfg.Headers, "Headers should be initialized correctly")
	assert.NotNil(t, hfg.index, "Index should be initialized")
	assert.True(t, hfg.IsHeaderInGroup("Authorization"), "Authorization header should be in the index")
	assert.True(t, hfg.IsHeaderInGroup("Cookie"), "Cookie header should be in the index")
	assert.False(t, hfg.IsHeaderInGroup("Non-Existent-Header"), "Non-Existent-Header should not be in the index")
}

func TestIsHeaderInGroup(t *testing.T) {
	t.Parallel()
	headers := []string{"Authorization", "Cookie"}
	hfg := NewHeaderFilterGroup(headers)

	assert.True(t, hfg.IsHeaderInGroup("Authorization"), "Authorization header should be in the index")
	assert.True(t, hfg.IsHeaderInGroup("Cookie"), "Cookie header should be in the index")
	assert.False(t, hfg.IsHeaderInGroup("Non-Existent-Header"), "Non-Existent-Header should not be in the index")
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
