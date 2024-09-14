package fileutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeSHA256FromFileContents(t *testing.T) {
	t.Run("ValidFile", func(t *testing.T) {
		content := []byte("Hello, World!")
		tmpFilePath := t.TempDir() + "/ValidFile.txt"
		tmpfile, err := os.Create(tmpFilePath)
		require.NoError(t, err)

		_, err = tmpfile.Write(content)
		require.NoError(t, err)

		err = tmpfile.Close()
		require.NoError(t, err)

		expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
		hash, err := ComputeSHA256FromFileContents(tmpFilePath)
		require.NoError(t, err)
		assert.Equal(t, expectedHash, hash)
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		tmpFilePath := t.TempDir() + "/NonExistentFile.txt"
		_, err := ComputeSHA256FromFileContents(tmpFilePath)
		assert.Error(t, err)
	})

	t.Run("EmptyFile", func(t *testing.T) {
		tmpFilePath := t.TempDir() + "/EmptyFile.txt"
		tmpfile, err := os.Create(tmpFilePath)
		require.NoError(t, err)

		err = tmpfile.Close()
		assert.NoError(t, err)

		expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		hash, err := ComputeSHA256FromFileContents(tmpFilePath)
		assert.NoError(t, err)
		assert.Equal(t, expectedHash, hash)
	})
}
