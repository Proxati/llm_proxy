package rest

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	requestTimeout           = 5
	headerID                 = "Llm_Proxy_Identifier"
	maxResponseBodyReadBytes = 1024 // 1KB
)

// EndpointSyncREST represents a single REST endpoint, and implements the Endpoint interface
type EndpointSyncREST struct {
	Name   string
	URL    string
	logger *slog.Logger
}

// NewEndpointSyncREST creates the EndpointSyncREST object, which implements the Endpoint interface
// it requires a slogger, a human-readable name, and a target REST URL
func NewEndpointSyncREST(logger *slog.Logger, name, url string) *EndpointSyncREST {
	return &EndpointSyncREST{
		Name:   name,
		URL:    url,
		logger: logger.WithGroup("EndpointAsyncREST"),
	}
}

// String returns the name of the endpoint
func (e *EndpointSyncREST) String() string {
	return fmt.Sprintf("AsyncREST: %s", e.Name)
}

// GetURL returns the URL of the endpoint
func (e *EndpointSyncREST) GetURL() string {
	return e.URL
}

// POST is a simple blocking POST request to a REST endpoint
func (e *EndpointSyncREST) POST(identifier string, data []byte) error {
	logger := e.logger.With("identifier", identifier)
	logger.Debug("POST'ing data", "endpoint", e.String(), "timeout", requestTimeout)

	// Create a new HTTP POST request
	req, err := http.NewRequest(http.MethodPost, e.GetURL(), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if identifier != "" {
		req.Header.Set(headerID, identifier)
	}

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: requestTimeout * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	limitedReader := io.LimitedReader{R: resp.Body, N: maxResponseBodyReadBytes}
	bodyBytes, err := io.ReadAll(&limitedReader)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	logger.Debug("Sent data to endpoint", "status", resp.Status, "body", string(bodyBytes))

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %d", resp.StatusCode)
	}
	return nil
}
