package fileUtils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelocateExistingFileIfExists(t *testing.T) {
	// Create a temporary directory for testing
	dir := t.TempDir()

	// Create a temporary file
	file, err := os.CreateTemp(dir, "test")
	assert.NoError(t, err)
	fileName := file.Name()

	// Close and remove the file
	assert.NoError(t, file.Close())
	assert.NoError(t, os.Remove(fileName))
	assert.False(t, FileExists(fileName))

	// Test with a file that does not exist
	err = RelocateExistingFileIfExists(fileName)
	require.NoError(t, err)

	// Create the file again
	file, err = os.Create(fileName)
	assert.NoError(t, err)
	file.Close()

	// Test with a file that does exist
	err = RelocateExistingFileIfExists(fileName)
	require.NoError(t, err)

	// Check that the original file does not exist
	assert.False(t, FileExists(fileName))

	// Check that a new file exists
	files, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
}
