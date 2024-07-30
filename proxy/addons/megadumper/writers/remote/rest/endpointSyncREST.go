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
	RequestTimeout           = 5
	ID_Header                = "Llm_Proxy_Identifier"
	MaxResponseBodyReadBytes = 1024 // 1KB
)

// Endpoint represents a single REST endpoint, and implements the Endpoint interface
type EndpointSyncREST struct {
	Name   string
	URL    string
	logger *slog.Logger
}

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

// PostData is a simple blocking POST to a REST endpoint
func (e *EndpointSyncREST) POST(identifier string, data []byte) error {
	logger := e.logger.With("identifier", identifier)
	logger.Debug("POST'ing data", "endpoint", e.String(), "timeout", RequestTimeout)

	// Create a new HTTP POST request
	req, err := http.NewRequest(http.MethodPost, e.GetURL(), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if identifier != "" {
		req.Header.Set(ID_Header, identifier)
	}

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: RequestTimeout * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	limitedReader := io.LimitedReader{R: resp.Body, N: MaxResponseBodyReadBytes}
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
