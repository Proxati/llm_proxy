package runners

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"log/slog"

	"github.com/proxati/llm_proxy/v2/internal/fileutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPosixCommand_Success(t *testing.T) {
	logger := slog.Default()
	command := "/bin/echo"
	concurrency := 0
	timeout := 5 * time.Second
	ctx := context.Background()

	// Compute the SHA256 checksum of the command
	sha256sum, err := fileutils.ComputeSHA256FromFileContents(command)
	assert.NoError(t, err)

	posixCmd, err := NewPosixCommand(logger, ctx, command, concurrency, timeout)
	assert.NoError(t, err)
	assert.NotNil(t, posixCmd)
	assert.Equal(t, sha256sum, posixCmd.sha256sum)
	assert.Equal(t, command, posixCmd.command)
	assert.Equal(t, concurrency, posixCmd.concurrency)
}

func TestNewPosixCommand_InvalidCommand(t *testing.T) {
	logger := slog.Default()
	command := "/invalid/command"
	concurrency := 0
	timeout := 5 * time.Second
	ctx := context.Background()

	// Compute the SHA256 checksum of the command
	_, err := fileutils.ComputeSHA256FromFileContents(command)
	assert.Error(t, err)

	posixCmd, err := NewPosixCommand(logger, ctx, command, concurrency, timeout)
	assert.Error(t, err)
	assert.Nil(t, posixCmd)
}

func TestNewPosixCommand_WithConcurrency(t *testing.T) {
	logger := slog.Default()
	command := "/bin/echo"
	concurrency := 5
	timeout := 5 * time.Second
	ctx := context.Background()

	// Compute the SHA256 checksum of the command
	sha256sum, err := fileutils.ComputeSHA256FromFileContents(command)
	assert.NoError(t, err)

	posixCmd, err := NewPosixCommand(logger, ctx, command, concurrency, timeout)
	assert.NoError(t, err)
	assert.NotNil(t, posixCmd)
	assert.Equal(t, sha256sum, posixCmd.sha256sum)
	assert.Equal(t, command, posixCmd.command)
	assert.Equal(t, concurrency, posixCmd.concurrency)
	assert.NotNil(t, posixCmd.bulkhead)
	assert.Equal(t, timeout, posixCmd.bulkheadTimeout)
}

func TestRun_Success(t *testing.T) {
	logger := slog.Default()
	concurrency := 0
	timeout := 5 * time.Second
	ctx := context.Background()
	stdin := bytes.NewReader([]byte("hello"))

	// Create a temporary directory
	tempDir := t.TempDir()

	// Define the script content
	scriptContent := `#!/bin/sh
cat`

	// Write the script to the temporary directory
	scriptPath := tempDir + "/test_script.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Use the script as the input to NewPosixCommand
	posixCmd, err := NewPosixCommand(logger, ctx, scriptPath, concurrency, timeout)
	assert.NoError(t, err)
	assert.NotNil(t, posixCmd)

	output, err := posixCmd.Run(ctx, stdin)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(output))
}

func TestRun_CommandExecutionFailure(t *testing.T) {
	logger := slog.Default()
	concurrency := 0
	timeout := 5 * time.Second
	ctx := context.Background()
	stdin := bytes.NewReader([]byte(""))

	// Create a temporary directory
	tempDir := t.TempDir()

	// Define the script content that will fail
	scriptContent := `#!/bin/sh
exit 1`

	// Write the script to the temporary directory
	scriptPath := tempDir + "/fail_script.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Use the script as the input to NewPosixCommand
	posixCmd, err := NewPosixCommand(logger, ctx, scriptPath, concurrency, timeout)
	assert.NoError(t, err)
	assert.NotNil(t, posixCmd)

	output, err := posixCmd.Run(ctx, stdin)
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestRun_BulkheadPermitFailure(t *testing.T) {
	logger := slog.Default()
	concurrency := 1
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	stdin := bytes.NewReader([]byte("hello"))

	// Create a temporary directory
	tempDir := t.TempDir()

	// Define a script that will block forever
	scriptContent := `#!/bin/sh
cat`

	// Write the script to the temporary directory
	scriptPath := tempDir + "/test_script.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	assert.NoError(t, err)

	// Change the permissions to make it executable
	err = os.Chmod(scriptPath, 0755)
	assert.NoError(t, err)

	// Use the script as the input to NewPosixCommand
	posixCmd, err := NewPosixCommand(logger, ctx, scriptPath, concurrency, timeout)
	assert.NoError(t, err)
	assert.NotNil(t, posixCmd)

	// Simulate bulkhead permit acquisition failure by running the command in a separate goroutine
	go func() {
		_, _ = posixCmd.Run(context.Background(), stdin)
	}()

	time.Sleep(10 * time.Millisecond) // Ensure the first command acquires the permit

	output, err := posixCmd.Run(ctx, stdin)
	assert.Error(t, err)
	assert.Nil(t, output)
}
