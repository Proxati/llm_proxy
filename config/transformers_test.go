package config

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseURL(t *testing.T, s string) url.URL {
	t.Helper()
	u, err := url.Parse(s)
	require.NoError(t, err)
	return *u
}

func TestNewTransformer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		expected *Transformer
	}{
		{
			name: "DefaultValues",
			expected: &Transformer{
				FailureMode:      FailureModeHard,
				Concurrency:      5,
				RequestTimeout:   1 * time.Second,
				ResponseTimeout:  60 * time.Second,
				RetryCount:       2,
				InitialRetryTime: 100 * time.Millisecond,
				MaxRetryTime:     60 * time.Second,
				// // BackPressureMode:  BackPressureModeNone,
				HealthCheck: TransformerHealthCheck{
					Interval: 30 * time.Second,
					Path:     "/health",
					Timeout:  1 * time.Second,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewTransformer()
			assert.Equal(t, tt.expected, transformer)
		})
	}
}

func TestNewTransformerWithInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected *Transformer
		hasError bool
	}{
		{
			name:  "ValidInput_AllOptions",
			input: "http://example.com/foo|timeout=10s,name=test,failure-mode=soft,concurrency=5,retry-count=3,initial-retry-time=2s,health-check.interval=1m,health-check.path=/status,health-check.timeout=5s,request-timeout=2s,initial-retry-time=500ms,max-retry-time=10s",
			expected: &Transformer{
				rawInput:         "http://example.com/foo|timeout=10s,name=test,failure-mode=soft,concurrency=5,retry-count=3,initial-retry-time=2s,health-check.interval=1m,health-check.path=/status,health-check.timeout=5s,request-timeout=2s,initial-retry-time=500ms,max-retry-time=10s",
				rawOptions:       "timeout=10s,name=test,failure-mode=soft,concurrency=5,retry-count=3,initial-retry-time=2s,health-check.interval=1m,health-check.path=/status,health-check.timeout=5s,request-timeout=2s,initial-retry-time=500ms,max-retry-time=10s",
				rawURL:           "http://example.com/foo",
				URL:              parseURL(t, "http://example.com/foo"),
				Name:             "test",
				FailureMode:      FailureModeSoft,
				Concurrency:      5,
				RequestTimeout:   2 * time.Second,
				ResponseTimeout:  10 * time.Second,
				RetryCount:       3,
				InitialRetryTime: 500 * time.Millisecond,
				MaxRetryTime:     10 * time.Second,
				// BackPressureMode:  BackPressureMode429,
				HealthCheck: TransformerHealthCheck{
					Interval: 1 * time.Minute,
					Path:     "/status",
					Timeout:  5 * time.Second,
				},
			},
			hasError: false,
		},
		{
			name:  "ValidInput_MinimumOptions",
			input: "grpc://example.com:1234/foo",
			expected: &Transformer{
				rawInput:         "grpc://example.com:1234/foo",
				rawOptions:       "",
				Name:             "grpc://example.com:1234/foo",
				rawURL:           "grpc://example.com:1234/foo",
				URL:              parseURL(t, "grpc://example.com:1234/foo"),
				FailureMode:      FailureModeHard,
				Concurrency:      5,
				RequestTimeout:   1 * time.Second,
				ResponseTimeout:  60 * time.Second,
				RetryCount:       2,
				InitialRetryTime: 100 * time.Millisecond,
				MaxRetryTime:     60 * time.Second,
				// BackPressureMode:  BackPressureModeNone,
				HealthCheck: TransformerHealthCheck{
					Interval: 30 * time.Second,
					Path:     "/health",
					Timeout:  1 * time.Second,
				},
			},
			hasError: false,
		},
		{
			name:     "InvalidInput_Empty",
			input:    "",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_MissingURL",
			input:    "|timeout=10s",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidOptionFormat",
			input:    "http://example.com/foo|timeout=10s,invalid-option",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidTimeout",
			input:    "http://example.com/foo|timeout=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidRetries",
			input:    "http://example.com/foo|retry-count=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidRetryInterval",
			input:    "http://example.com/foo|initial-retry-time=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidConcurrency",
			input:    "http://example.com/foo|concurrency=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidHealthCheckInterval",
			input:    "http://example.com/foo|health-check.interval=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidHealthCheckPath",
			input:    "http://example.com/foo|health-check.path=status",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidFailMode",
			input:    "http://example.com/foo|failure-mode=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_TimeoutTooShort",
			input:    "http://example.com/foo|timeout=500us",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_RetryIntervalTooShort",
			input:    "http://example.com/foo|initial-retry-time=500us",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_ConcurrencyTooLow",
			input:    "http://example.com/foo|concurrency=0",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_HealthCheckPathWithoutSlash",
			input:    "http://example.com/foo|health-check.path=status",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_UnknownKey",
			input:    "http://example.com/foo|blarg=zarp",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_MaxRetryTimeLessThanInitialRetryTime",
			input:    "http://example.com/foo|max-retry-time=1s,initial-retry-time=2s",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidRetryCount",
			input:    "http://example.com/foo|retry-count=-1",
			expected: nil,
			hasError: true,
		},
		/*
			{
				name:     "InvalidInput_InvalidBackPressureMode",
				input:    "http://example.com/foo|back-pressure-mode=invalid",
				expected: nil,
				hasError: true,
			},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer, err := NewTransformerWithInput(tt.input)
			if tt.hasError {
				require.Error(t, err)
				assert.Nil(t, transformer)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, transformer)
			}
		})
	}
}

func TestSetInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected *Transformer
		hasError bool
	}{
		{
			name:  "ValidInput_AllOptions",
			input: "http://example.com/foo|timeout=10s,name=test,failure-mode=soft,concurrency=5,retry-count=3,initial-retry-time=2s,health-check.interval=1m,health-check.path=/status,health-check.timeout=5s,request-timeout=2s,initial-retry-time=500ms,max-retry-time=10s",
			expected: &Transformer{
				rawInput:         "http://example.com/foo|timeout=10s,name=test,failure-mode=soft,concurrency=5,retry-count=3,initial-retry-time=2s,health-check.interval=1m,health-check.path=/status,health-check.timeout=5s,request-timeout=2s,initial-retry-time=500ms,max-retry-time=10s",
				rawOptions:       "timeout=10s,name=test,failure-mode=soft,concurrency=5,retry-count=3,initial-retry-time=2s,health-check.interval=1m,health-check.path=/status,health-check.timeout=5s,request-timeout=2s,initial-retry-time=500ms,max-retry-time=10s",
				Name:             "test",
				rawURL:           "http://example.com/foo",
				URL:              parseURL(t, "http://example.com/foo"),
				FailureMode:      FailureModeSoft,
				Concurrency:      5,
				RequestTimeout:   2 * time.Second,
				ResponseTimeout:  10 * time.Second,
				RetryCount:       3,
				InitialRetryTime: 500 * time.Millisecond,
				MaxRetryTime:     10 * time.Second,
				// BackPressureMode:  BackPressureMode429,
				HealthCheck: TransformerHealthCheck{
					Interval: 1 * time.Minute,
					Path:     "/status",
					Timeout:  5 * time.Second,
				},
			},
			hasError: false,
		},
		{
			name:  "ValidInput_MinimumOptions",
			input: "grpc://example.com:1234/foo",
			expected: &Transformer{
				rawInput:         "grpc://example.com:1234/foo",
				rawOptions:       "",
				Name:             "grpc://example.com:1234/foo",
				rawURL:           "grpc://example.com:1234/foo",
				URL:              parseURL(t, "grpc://example.com:1234/foo"),
				FailureMode:      FailureModeHard,
				Concurrency:      5,
				RequestTimeout:   1 * time.Second,
				ResponseTimeout:  60 * time.Second,
				RetryCount:       2,
				InitialRetryTime: 100 * time.Millisecond,
				MaxRetryTime:     60 * time.Second,
				// BackPressureMode:  BackPressureModeNone,
				HealthCheck: TransformerHealthCheck{
					Interval: 30 * time.Second,
					Path:     "/health",
					Timeout:  1 * time.Second,
				},
			},
			hasError: false,
		},
		{
			name:     "InvalidInput_Empty",
			input:    "",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_MissingURL",
			input:    "|timeout=10s",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidOptionFormat",
			input:    "http://example.com/foo|timeout=10s,invalid-option",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidTimeout",
			input:    "http://example.com/foo|timeout=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidRetries",
			input:    "http://example.com/foo|retry-count=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidRetryInterval",
			input:    "http://example.com/foo|initial-retry-time=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidConcurrency",
			input:    "http://example.com/foo|concurrency=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidHealthCheckInterval",
			input:    "http://example.com/foo|health-check.interval=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidHealthCheckPath",
			input:    "http://example.com/foo|health-check.path=status",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_InvalidFailMode",
			input:    "http://example.com/foo|failure-mode=invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_TimeoutTooShort",
			input:    "http://example.com/foo|timeout=500us",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_RetryIntervalTooShort",
			input:    "http://example.com/foo|initial-retry-time=500us",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_ConcurrencyTooLow",
			input:    "http://example.com/foo|concurrency=0",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_HealthCheckPathWithoutSlash",
			input:    "http://example.com/foo|health-check.path=status",
			expected: nil,
			hasError: true,
		},
		{
			name:     "InvalidInput_UnknownKey",
			input:    "http://example.com/foo|blarg=zarp",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewTransformer()
			err := transformer.SetInput(tt.input)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, transformer)
			}
		})
	}
}

func TestNewTrafficTransformers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		requestTransformerInput  string
		responseTransformerInput string
		expectedRequest          []*Transformer
		expectedResponse         []*Transformer
		expectError              bool
	}{
		{
			name:                     "ValidInput_SingleRequestTransformer",
			requestTransformerInput:  "http://example.com/request|timeout=10s,name=request-transformer",
			responseTransformerInput: "",
			expectedRequest: []*Transformer{
				{
					rawInput:         "http://example.com/request|timeout=10s,name=request-transformer",
					rawOptions:       "timeout=10s,name=request-transformer",
					rawURL:           "http://example.com/request",
					URL:              parseURL(t, "http://example.com/request"),
					Name:             "request-transformer",
					FailureMode:      FailureModeHard,
					Concurrency:      5,
					RequestTimeout:   1 * time.Second,
					ResponseTimeout:  10 * time.Second,
					RetryCount:       2,
					InitialRetryTime: 100 * time.Millisecond,
					MaxRetryTime:     60 * time.Second,
					// BackPressureMode:  BackPressureModeNone,
					HealthCheck: TransformerHealthCheck{
						Interval: 30 * time.Second,
						Path:     "/health",
						Timeout:  1 * time.Second,
					},
				},
			},
			expectedResponse: []*Transformer{},
			expectError:      false,
		},
		{
			name:                     "ValidInput_SingleResponseTransformer",
			requestTransformerInput:  "",
			responseTransformerInput: "http://example.com/response|timeout=10s,name=response-transformer",
			expectedRequest:          []*Transformer{},
			expectedResponse: []*Transformer{
				{
					rawInput:         "http://example.com/response|timeout=10s,name=response-transformer",
					rawOptions:       "timeout=10s,name=response-transformer",
					rawURL:           "http://example.com/response",
					URL:              parseURL(t, "http://example.com/response"),
					Name:             "response-transformer",
					FailureMode:      FailureModeHard,
					Concurrency:      5,
					RequestTimeout:   1 * time.Second,
					ResponseTimeout:  10 * time.Second,
					RetryCount:       2,
					InitialRetryTime: 100 * time.Millisecond,
					MaxRetryTime:     60 * time.Second,
					// BackPressureMode:  BackPressureModeNone,
					HealthCheck: TransformerHealthCheck{
						Interval: 30 * time.Second,
						Path:     "/health",
						Timeout:  1 * time.Second,
					},
				},
			},
			expectError: false,
		},
		{
			name:                     "ValidInput_MultipleRequestTransformers",
			requestTransformerInput:  "http://example.com/request1|timeout=10s,name=request-transformer1;http://example.com/request2|timeout=20s,name=request-transformer2",
			responseTransformerInput: "",
			expectedRequest: []*Transformer{
				{
					rawInput:         "http://example.com/request1|timeout=10s,name=request-transformer1",
					rawOptions:       "timeout=10s,name=request-transformer1",
					rawURL:           "http://example.com/request1",
					URL:              parseURL(t, "http://example.com/request1"),
					Name:             "request-transformer1",
					FailureMode:      FailureModeHard,
					Concurrency:      5,
					RequestTimeout:   1 * time.Second,
					ResponseTimeout:  10 * time.Second,
					RetryCount:       2,
					InitialRetryTime: 100 * time.Millisecond,
					MaxRetryTime:     60 * time.Second,
					// BackPressureMode:  BackPressureModeNone,
					HealthCheck: TransformerHealthCheck{
						Interval: 30 * time.Second,
						Path:     "/health",
						Timeout:  1 * time.Second,
					},
				},
				{
					rawInput:         "http://example.com/request2|timeout=20s,name=request-transformer2",
					rawOptions:       "timeout=20s,name=request-transformer2",
					rawURL:           "http://example.com/request2",
					URL:              parseURL(t, "http://example.com/request2"),
					Name:             "request-transformer2",
					FailureMode:      FailureModeHard,
					Concurrency:      5,
					RequestTimeout:   1 * time.Second,
					ResponseTimeout:  20 * time.Second,
					RetryCount:       2,
					InitialRetryTime: 100 * time.Millisecond,
					MaxRetryTime:     60 * time.Second,
					// BackPressureMode:  BackPressureModeNone,
					HealthCheck: TransformerHealthCheck{
						Interval: 30 * time.Second,
						Path:     "/health",
						Timeout:  1 * time.Second,
					},
				},
			},
			expectedResponse: []*Transformer{},
			expectError:      false,
		},
		{
			name:                     "InvalidInput_EmptyRequestTransformer",
			requestTransformerInput:  "",
			responseTransformerInput: "",
			expectedRequest:          []*Transformer{},
			expectedResponse:         []*Transformer{},
			expectError:              false,
		},
		{
			name:                     "InvalidInput_InvalidRequestTransformer",
			requestTransformerInput:  "http://example.com/request|timeout=invalid",
			responseTransformerInput: "",
			expectedRequest:          nil,
			expectedResponse:         nil,
			expectError:              true,
		},
		{
			name:                     "InvalidInput_InvalidResponseTransformer",
			requestTransformerInput:  "",
			responseTransformerInput: "http://example.com/response|timeout=invalid",
			expectedRequest:          nil,
			expectedResponse:         nil,
			expectError:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformers, err := NewTrafficTransformers(tt.requestTransformerInput, tt.responseTransformerInput)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRequest, transformers.Request)
				assert.Equal(t, tt.expectedResponse, transformers.Response)
			}
		})
	}
}
