package fileutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExecutable_FileDoesNotExist(t *testing.T) {
	t.Parallel()
	err := IsExecutable("/path/to/nonexistent/file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to stat file")
}

func TestIsExecutable_FileNotExecutable(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "not_executable")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Ensure the file is not executable
	err = os.Chmod(tmpFile.Name(), 0644)
	assert.NoError(t, err)

	err = IsExecutable(tmpFile.Name())
	assert.Error(t, err)
	assert.Equal(t, "command is not executable", err.Error())
}

func TestIsExecutable_FileIsExecutable(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "executable")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Make the file executable
	err = os.Chmod(tmpFile.Name(), 0755)
	assert.NoError(t, err)

	// Check if the file is executable
	err = IsExecutable(tmpFile.Name())
	assert.NoError(t, err)
}
