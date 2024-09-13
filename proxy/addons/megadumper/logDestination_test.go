package megadumper

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"log/slog"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogDestinations(t *testing.T) {
	logger := slog.Default()
	invalidPath := `/c:\/../*^`

	t.Run("Empty logTarget defaults to stdout", func(t *testing.T) {
		configs, err := NewLogDestinations(logger, "", config.LogFormatJSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, "stdout", configs[0].target)
	})

	t.Run("Valid file path with file:// prefix creates writer for directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		configs, err := NewLogDestinations(logger, "file://"+tmpDir, config.LogFormatJSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, tmpDir, configs[0].target)
	})

	t.Run("Valid file path without file:// prefix creates writer for directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		configs, err := NewLogDestinations(logger, tmpDir, config.LogFormatJSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, tmpDir, configs[0].target)
	})

	t.Run("Valid file path with http:// prefix creates writer for an asyncREST", func(t *testing.T) {
		exampleURL := "http://example.com"

		configs, err := NewLogDestinations(logger, exampleURL, config.LogFormatJSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, exampleURL, configs[0].target)
	})

	t.Run("Valid file path with https:// prefix creates writer for an asyncREST", func(t *testing.T) {
		exampleURL := "https://example.com"

		configs, err := NewLogDestinations(logger, exampleURL, config.LogFormatJSON)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, exampleURL, configs[0].target)
	})

	t.Run("Valid file paths with a mix of file:// http:// prefixes creates multiple writers", func(t *testing.T) {
		exampleURL := "https://example.com"
		tmpDir := t.TempDir()

		configs, err := NewLogDestinations(logger, exampleURL+","+tmpDir, config.LogFormatJSON)
		require.NoError(t, err)
		require.Len(t, configs, 2)
		assert.Equal(t, exampleURL, configs[0].target)
		assert.Equal(t, tmpDir, configs[1].target)
	})

	t.Run("Multiple file paths without file:// prefix creates writers for directory", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		configs, err := NewLogDestinations(logger, tmpDir1+","+tmpDir2, config.LogFormatJSON)
		require.NoError(t, err)
		require.Len(t, configs, 2)
		assert.Equal(t, tmpDir1, configs[0].target)
		assert.Equal(t, tmpDir2, configs[1].target)
	})

	t.Run("Invalid file path returns error", func(t *testing.T) {
		_, err := NewLogDestinations(logger, invalidPath, config.LogFormatJSON)
		require.Error(t, err)
	})

	t.Run("Multiple valid and invalid targets", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		configs, err := NewLogDestinations(logger, fmt.Sprintf("file://%s,%s,%s", tmpDir1, invalidPath, tmpDir2), config.LogFormatJSON)
		require.Error(t, err)
		require.Nil(t, configs)
	})
}

func TestLogDestination_Write(t *testing.T) {
	t.Parallel()
	logger := slog.Default()
	tmpDir := t.TempDir()
	logTarget := "file://" + tmpDir
	format := config.LogFormatJSON

	// Create LogDestination
	logDestinations, err := NewLogDestinations(logger, logTarget, format)
	require.NoError(t, err)
	require.Len(t, logDestinations, 1)
	logDestination := logDestinations[0]

	// Create LogDumpContainer
	logDumpContainer := &schema.LogDumpContainer{
		ObjectType:    "testing",
		SchemaVersion: "99",
		Timestamp:     time.Now(),
	}

	// Write LogDumpContainer
	identifier := "test_identifier"
	_, err = logDestination.Write(identifier, *logDumpContainer)
	require.NoError(t, err)

	// Verify Output
	expectedFilePath := fmt.Sprintf("%s/%s.json", tmpDir, identifier)
	assert.FileExists(t, expectedFilePath)

	// read the file and verify its contents
	content, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	expectedJSON, err := json.Marshal(logDumpContainer)
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(content))
}
