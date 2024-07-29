package fileUtils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	assert.False(t, FileExists(dir+"/nonexistent.txt"))

	_, err := os.Create(dir + "/existent.txt")
	require.NoError(t, err)

	assert.True(t, FileExists(dir+"/existent.txt"))
}
