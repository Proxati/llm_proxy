package writers_test

import (
	"os"
	"testing"

	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/formatters"
	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/writers"
	"github.com/stretchr/testify/assert"
)

func TestToDir_Write(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // clean up

	// Create a new ToDir instance with a pointer to PlainText to satisfy the interface
	toDir, err := writers.NewToDir(tempDir, &formatters.PlainText{})
	assert.NoError(t, err)

	// Write some data
	data := []byte("test data")
	n, err := toDir.Write("test", data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Check that the file was created and contains the correct data
	files, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	fileData, err := os.ReadFile(tempDir + "/" + files[0].Name())
	assert.NoError(t, err)
	assert.Equal(t, data, fileData)
}
