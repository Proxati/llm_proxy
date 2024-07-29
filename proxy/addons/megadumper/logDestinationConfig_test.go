package megadumper

import (
	"fmt"
	"testing"

	"log/slog"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogDestinationConfigs(t *testing.T) {
	t.Parallel()
	logger := slog.Default()
	invalidPath := `/c:\/../*^`

	t.Run("Empty logTarget defaults to stdout", func(t *testing.T) {
		configs, err := NewLogDestinationConfigs(logger, "", config.LogFormat_JSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, "stdout", configs[0].target)
	})

	t.Run("Valid file path with file:// prefix creates writer for directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		configs, err := NewLogDestinationConfigs(logger, "file://"+tmpDir, config.LogFormat_JSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, tmpDir, configs[0].target)
	})

	t.Run("Valid file path without file:// prefix creates writer for directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		configs, err := NewLogDestinationConfigs(logger, tmpDir, config.LogFormat_JSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, tmpDir, configs[0].target)
	})

	t.Run("Multiple file paths without file:// prefix creates writers for directory", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		configs, err := NewLogDestinationConfigs(logger, tmpDir1+","+tmpDir2, config.LogFormat_JSON)
		require.NoError(t, err)
		require.Len(t, configs, 2)
		assert.Equal(t, tmpDir1, configs[0].target)
		assert.Equal(t, tmpDir2, configs[1].target)
	})

	t.Run("Invalid file path returns error", func(t *testing.T) {
		_, err := NewLogDestinationConfigs(logger, invalidPath, config.LogFormat_JSON)
		require.Error(t, err)
	})

	t.Run("Multiple valid and invalid targets", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		configs, err := NewLogDestinationConfigs(logger, fmt.Sprintf("file://%s,%s,%s", tmpDir1, invalidPath, tmpDir2), config.LogFormat_JSON)
		require.Error(t, err)
		require.Nil(t, configs)
	})
}
