package fileutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirExistsOrCreate(t *testing.T) {
	dir := t.TempDir()

	err := DirExistsOrCreate(dir + "/subdir")
	require.NoError(t, err)

	_, err = os.Stat(dir + "/subdir")
	assert.NoError(t, err)
}
