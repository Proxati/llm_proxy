package memory_Engine

import (
	"testing"

	"github.com/proxati/llm_proxy/v2/proxy/addons/cache/key"
	"github.com/stretchr/testify/require"
)

func TestSetGetBytes(t *testing.T) {
	m, err := NewMemoryStorage(10)
	require.NoError(t, err)

	key := key.NewKeyStr("key") // replace with actual key structure if different
	value := []byte("value")

	err = m.SetBytes("testIdentifier", key, value)
	require.NoError(t, err)

	retrievedValue, err := m.GetBytes("testIdentifier", key)
	require.NoError(t, err)
	require.Equal(t, value, retrievedValue)
}
