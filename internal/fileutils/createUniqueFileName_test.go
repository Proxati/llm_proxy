package fileutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUniqueFileName(t *testing.T) {
	dir := t.TempDir()
	fileName := CreateUniqueFileName(dir, "test", "txt", 0)
	assert.Equal(t, fileName, dir+"/test.txt")

	// Create a file to test the unique filename generation
	_, err := os.Create(fileName)
	require.NoError(t, err)

	fileName2 := CreateUniqueFileName(dir, "test", "txt", 0)
	assert.Equal(t, fileName2, dir+"/test-1.txt")
}
