package fileUtils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirExistsOrCreate(t *testing.T) {
	dir := t.TempDir()

	err := DirExistsOrCreate(dir + "/subdir")
	assert.NoError(t, err)

	_, err = os.Stat(dir + "/subdir")
	assert.NoError(t, err)
}
