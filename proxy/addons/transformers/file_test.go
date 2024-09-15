package transformers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"testing"

	"log/slog"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons/transformers/runners"
	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFileProvider_Transform(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Define the script content
	scriptContent := `#!/bin/sh
cat`

	// Write the script to the temporary directory
	scriptPath := tempDir + "/test_script.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err, "Failed to write test script")

	// Parse the script path as a URL with "file://" scheme
	u, err := url.Parse("file://" + scriptPath)
	require.NoError(t, err, "Failed to parse script path as URL")

	transformerConfig := &config.Transformer{
		Name:           "test-transformer",
		URL:            *u,
		Concurrency:    1,
		RequestTimeout: 1000,
	}

	fileProvider, err := NewFileProvider(logger, ctx, transformerConfig)
	require.NoError(t, err, "Failed to create FileProvider")

	// Helper function to create a LogDumpContainer with necessary fields
	createLogDumpContainer := func(reqBody string, reqHeader http.Header, respBody string, respHeader http.Header) (*schema.LogDumpContainer, error) {
		ldc := schema.NewLogDumpContainerEmpty()
		ldc.ObjectType = schema.ObjectTypeDefault  // Ensure ObjectType is set
		ldc.SchemaVersion = schema.SchemaVersionV2 // Ensure Schema is set
		ldc.Request = &schema.ProxyRequest{
			Body:   reqBody,
			Header: reqHeader,
		}
		ldc.Response = &schema.ProxyResponse{
			Body:   respBody,
			Header: respHeader,
		}
		return ldc, nil
	}

	t.Run("successful transformation with multiple fields updated", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		oldReq := &schema.ProxyRequest{
			Body:   "old-request",
			Header: http.Header{"X-Old": {"old-value"}},
		}
		oldResp := &schema.ProxyResponse{
			Body:   "old-response",
			Header: http.Header{"X-Old-Resp": {"old-value-resp"}},
		}

		newReq := &schema.ProxyRequest{
			Body:   "new-request",
			Header: http.Header{"X-New": {"new-value"}},
		}
		newResp := &schema.ProxyResponse{
			Body:   "new-response",
			Header: http.Header{"X-New-Resp": {"new-value-resp"}},
		}

		// Create a log dump container with the expected JSON output after calling .Run()
		expectedLDC, err := createLogDumpContainer(newReq.Body, newReq.Header, newResp.Body, newResp.Header)
		require.NoError(t, err, "Failed to create expected LogDumpContainer")

		expectedJson, err := expectedLDC.MarshalJSON()
		require.NoError(t, err, "Failed to marshal expected LogDumpContainer to JSON")

		t.Logf("Expected JSON: %s", string(expectedJson)) // Debug log

		// Set up mock expectations
		mockRunner.On("Run", mock.Anything, mock.AnythingOfType("*bytes.Reader")).Return(expectedJson, nil)

		req, resp, err := fileProvider.Transform(ctx, logger, oldReq, oldResp, nil, nil)
		require.NoError(t, err, "Transform should not return an error")

		// Verify that newReq and newResp are updated correctly
		assert.Equal(t, newReq.Body, req.Body, "Request body should be updated")
		assert.Equal(t, newReq.Header, req.Header, "Request headers should be updated")
		assert.Equal(t, newResp.Body, resp.Body, "Response body should be updated")
		assert.Equal(t, newResp.Header, resp.Header, "Response headers should be updated")

		mockRunner.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		_, _, err := fileProvider.Transform(ctx, logger, nil, nil, nil, nil)
		require.Error(t, err, "Transform should return an error when request is nil")
		assert.EqualError(t, err, "unable to transform, request object is nil")

		mockRunner.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
	})

	t.Run("command runner error", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		oldReq := &schema.ProxyRequest{Body: "old-request"}
		oldResp := &schema.ProxyResponse{Body: "old-response"}

		// Set up mock to return an error
		mockRunner.On("Run", mock.Anything, mock.AnythingOfType("*bytes.Reader")).Return(nil, errors.New("command runner error"))

		_, _, err := fileProvider.Transform(ctx, logger, oldReq, oldResp, nil, nil)
		require.Error(t, err, "Transform should return an error when command runner fails")
		assert.EqualError(t, err, "unable to run command: command runner error")

		mockRunner.AssertExpectations(t)
	})

	t.Run("invalid JSON returned by command runner", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		oldReq := &schema.ProxyRequest{Body: "old-request"}
		oldResp := &schema.ProxyResponse{Body: "old-response"}

		invalidJson := []byte("{invalid-json}") // Malformed JSON

		// Set up mock to return invalid JSON
		mockRunner.On("Run", mock.Anything, mock.AnythingOfType("*bytes.Reader")).Return(invalidJson, nil)

		_, _, err := fileProvider.Transform(ctx, logger, oldReq, oldResp, nil, nil)
		require.Error(t, err, "Transform should return an error when invalid JSON is returned")
		assert.Contains(t, err.Error(), "unable to unmarshal LogDumpContainer")

		mockRunner.AssertExpectations(t)
	})

	t.Run("empty JSON object", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		oldReq := &schema.ProxyRequest{Body: "old-request"}
		oldResp := &schema.ProxyResponse{Body: "old-response"}

		emptyJson := []byte(`{}`)

		// Set up mock to return empty JSON
		mockRunner.On("Run", mock.Anything, mock.AnythingOfType("*bytes.Reader")).Return(emptyJson, nil)

		_, _, err := fileProvider.Transform(ctx, logger, oldReq, oldResp, nil, nil)
		require.Error(t, err, "Transform should return an error when JSON is empty")
		assert.Contains(t, err.Error(), "object_type is required")

		mockRunner.AssertExpectations(t)
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		oldReq := &schema.ProxyRequest{Body: "old-request"}
		oldResp := &schema.ProxyResponse{Body: "old-response"}

		// Create a context that is already canceled
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		// Set up mock to simulate context cancellation
		mockRunner.On("Run", mock.Anything, mock.AnythingOfType("*bytes.Reader")).Return(nil, context.Canceled)

		_, _, err := fileProvider.Transform(canceledCtx, logger, oldReq, oldResp, nil, nil)
		require.Error(t, err, "Transform should return an error when context is canceled")
		assert.True(t, errors.Is(err, context.Canceled), "Error should be context.Canceled")

		mockRunner.AssertExpectations(t)
	})

	t.Run("health check success", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		// Set up mock for HealthCheck
		mockRunner.On("HealthCheck").Return(nil)

		err := fileProvider.HealthCheck(ctx)
		require.NoError(t, err, "HealthCheck should not return an error when runner is healthy")

		mockRunner.AssertExpectations(t)
	})

	t.Run("health check failure", func(t *testing.T) {
		mockRunner := new(runners.MockCommandRunner)
		fileProvider.commandRunner = mockRunner

		// Set up mock for HealthCheck to return an error
		mockRunner.On("HealthCheck").Return(errors.New("health check failed"))

		err := fileProvider.HealthCheck(ctx)
		require.Error(t, err, "HealthCheck should return an error when runner is unhealthy")
		assert.EqualError(t, err, "health check failed")

		mockRunner.AssertExpectations(t)
	})
}
