package fileUtils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNewFileFromFilename(t *testing.T) {
	dir := t.TempDir()

	file, err := CreateNewFileFromFilename(dir + "/test.txt")
	require.NoError(t, err)
	defer file.Close()

	_, err = os.Stat(dir + "/test.txt")
	assert.NoError(t, err)
}
