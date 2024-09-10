package config

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// FailureMode is a custom type representing the failure mode of a transformer
type FailureMode string

// BackPressureMode is a custom type representing the back pressure on clients when a transformer
// is unavailable, responding with 429s, or timeouts
type BackPressureMode string

const (
	// FailureModeHard means respond to the client with a 500 if a transformation is unsuccessful
	FailureModeHard FailureMode = "hard"

	// FailureModeSoft will allow skipping a transformation if the transformer is unavailable
	FailureModeSoft FailureMode = "soft"

	// BackPressureModeNone means no back pressure is applied
	BackPressureModeNone BackPressureMode = "none"

	// BackPressureMode429 means respond with a 429 if a transformer is unavailable
	BackPressureMode429 BackPressureMode = "429"

	// BackPressureModeTimeout means respond with a 504 if a transformer times out
	BackPressureModeTimeout BackPressureMode = "timeout"

	transformerTagName = "transformer" // used by the tag parser, keep it simple to match for field.Tag.Lookup()
)

// TransformerHealthCheck represents the health check configuration for a transformer
// Fields:
// - Interval: the interval at which to check the health of the transformer, -1 or 0 to disable
// - Path: the URL path to check for health
// - Timeout: the timeout for the health check
type TransformerHealthCheck struct {
	Interval time.Duration `transformer:"interval"`
	Path     string        `transformer:"path"`
	Timeout  time.Duration `transformer:"timeout"`
}

func (thc *TransformerHealthCheck) String() string {
	return fmt.Sprintf("TransformerHealthCheck{Interval: %s, Path: %s, Timeout: %s}",
		thc.Interval, thc.Path, thc.Timeout)
}

func (thc *TransformerHealthCheck) IsDisabled() bool {
	return thc.Interval <= 0
}

func (thc *TransformerHealthCheck) validate() error {
	var errs []error
	if thc.Timeout < time.Duration(1*time.Second) {
		errs = append(errs, errors.New("health check timeout must be at least 1 second"))
	}
	if thc.Path == "" {
		errs = append(errs, errors.New("empty health check path"))
	}
	if !strings.HasPrefix(thc.Path, "/") {
		errs = append(errs, errors.New("health check path must start with /"))
	}

	return errors.Join(errs...)
}

// Transformer represents a transformer configuration
// Fields:
// - URL: the URL of the transformer service
// - Name: the name of the transformer
// - FailureMode: the failure mode of the transformer (hard or soft)
// - Concurrency: the number of concurrent requests allowed to the transformer
// - RequestTimeout: the timeout for the request to be received by the transformer
// - ResponseTimeout: once the request has been received, the timeout for the transformer to respond
// - RetryCount: the number of retries to attempt if the transformer fails (0 for no retries)
// - InitialRetryTime: the initial time to wait before retrying a failed transformer
// - MaxRetryTime: the maximum time to wait before retrying a failed transformer (exponential backoff is used)
// - BackPressureMode: the back pressure mode to apply when the transformer is unavailable
// - HealthCheck: the health check configuration for the transformer
type Transformer struct {
	rawInput         string                 `transformer:"-"`
	rawOptions       string                 `transformer:"-"`
	rawURL           string                 `transformer:"-"`
	URL              url.URL                `transformer:"-"`
	Name             string                 `transformer:"name"`
	FailureMode      FailureMode            `transformer:"failure-mode"`
	Concurrency      int                    `transformer:"concurrency"`
	RequestTimeout   time.Duration          `transformer:"request-timeout"`
	ResponseTimeout  time.Duration          `transformer:"timeout"`
	RetryCount       int                    `transformer:"retry-count"`
	InitialRetryTime time.Duration          `transformer:"initial-retry-time"`
	MaxRetryTime     time.Duration          `transformer:"max-retry-time"`
	BackPressureMode BackPressureMode       `transformer:"back-pressure-mode"`
	HealthCheck      TransformerHealthCheck `transformer:"health-check"`
}

func NewTransformer() *Transformer {
	return &Transformer{
		FailureMode:      FailureModeHard,
		Concurrency:      5,
		RequestTimeout:   1 * time.Second,
		ResponseTimeout:  60 * time.Second,
		RetryCount:       2,
		InitialRetryTime: 100 * time.Millisecond,
		MaxRetryTime:     60 * time.Second,
		BackPressureMode: BackPressureModeNone,
		HealthCheck: TransformerHealthCheck{
			Interval: 30 * time.Second,
			Path:     "/health",
			Timeout:  1 * time.Second,
		},
	}
}

func NewTransformerWithInput(input string) (*Transformer, error) {
	output := NewTransformer()
	err := output.SetInput(input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (t Transformer) String() string {
	return fmt.Sprintf(
		"Transformer{ServiceName: %s, URL: %s, FailureMode: %s, InitialTimeout: %s, MaxRetries: %d, RetryInterval: %s, Concurrency: %d}",
		t.Name, t.rawURL, t.FailureMode, t.ResponseTimeout, t.RetryCount, t.InitialRetryTime, t.Concurrency,
	)
}

func (t *Transformer) validate() error {
	var errs []error

	hcErr := t.HealthCheck.validate()
	if hcErr != nil {
		errs = append(errs, hcErr)
	}

	if t.Name == "" {
		errs = append(errs, errors.New("empty transformer name"))
	}
	if t.URL.String() == "" {
		errs = append(errs, errors.New("empty transformer URL"))
	}
	if t.ResponseTimeout <= time.Duration(1*time.Millisecond) {
		errs = append(errs, errors.New("invalid response timeout"))
	}
	if t.RetryCount < 0 {
		errs = append(errs, errors.New("invalid retry count"))
	}
	if t.InitialRetryTime <= time.Duration(1*time.Millisecond) {
		errs = append(errs, errors.New("invalid initial retry"))
	}
	if t.MaxRetryTime < t.InitialRetryTime {
		errs = append(errs, errors.New("max-retry must be larger than initial-retry"))
	}
	if t.Concurrency <= 0 {
		errs = append(errs, errors.New("invalid concurrency"))
	}

	switch t.BackPressureMode {
	case BackPressureModeNone, BackPressureMode429, BackPressureModeTimeout:
		// Valid modes, do nothing
	default:
		errs = append(errs, errors.New("invalid back pressure mode"))
	}

	switch t.FailureMode {
	case FailureModeHard, FailureModeSoft:
		// Valid modes, do nothing
	default:
		errs = append(errs, errors.New("invalid failure mode"))
	}

	return errors.Join(errs...)
}

func (t *Transformer) SetInput(input string) error {
	if input == "" {
		return errors.New("empty transformer configuration")
	}

	t.rawInput = input

	rawURL, optsRaw, _ := strings.Cut(input, "|")
	if rawURL == "" {
		return errors.New("empty URL in transformer config: " + input)
	}

	// store a copy of the raw URL for debugging
	t.rawURL = rawURL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	t.URL = *parsedURL
	t.Name = rawURL
	if optsRaw == "" {
		// no options, return early
		return nil
	}

	// store a copy of the options for debugging
	t.rawOptions = optsRaw

	// chop the options into a map
	optsMap := make(map[string]string)
	for _, opt := range strings.Split(optsRaw, ",") {
		key, value, _ := strings.Cut(opt, "=")
		optsMap[key] = value
	}

	err = parseStructTags(transformerTagName, optsMap, reflect.ValueOf(t))
	if err != nil {
		return fmt.Errorf("invalid transformer options: %w", err)
	}

	return t.validate()
}

func (t *Transformer) GetHealthCheckURL() (string, error) {
	if t.HealthCheck.IsDisabled() {
		return "", nil
	}

	u, err := url.Parse(t.URL.String() + t.HealthCheck.Path)
	if err != nil {
		return "", fmt.Errorf("invalid health check URL: %w", err)
	}
	return u.String(), nil
}

type TrafficTransformers struct {
	Request  []*Transformer
	Response []*Transformer
}

// NewTrafficTransformers parses request and response transformation configurations
// into Transformer objects. Each configuration string contains one or more URLs
// with associated options, separated by semicolons (;).
//
// If any errors occur during parsing, they are collected and returned as a single error.
// On success, a TrafficTransformers struct with the parsed transformers is returned.
//
// Example:
//
//	requestTransformerConfig := "http://service1/request-filter|timeout=3000ms,fail-mode=soft;http://service2/request-filter|timeout=2000ms,fail-mode=hard"
//	responseTransformerConfig := "http://service2/response-filter|timeout=5000ms,fail-mode=hard"
func NewTrafficTransformers(requestTransformerInput string, responseTransformerInput string) (*TrafficTransformers, error) {
	var errs []error
	tt := &TrafficTransformers{
		Request:  make([]*Transformer, 0),
		Response: make([]*Transformer, 0),
	}

	if requestTransformerInput != "" {
		for _, config := range strings.Split(requestTransformerInput, ";") {
			transformer, err := NewTransformerWithInput(config)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if transformer == nil {
				errs = append(errs, errors.New("nil transformer from config: "+config))
				continue
			}
			tt.Request = append(tt.Request, transformer)
		}
	}

	if responseTransformerInput != "" {
		for _, config := range strings.Split(responseTransformerInput, ";") {
			transformer, err := NewTransformerWithInput(config)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if transformer == nil {
				errs = append(errs, errors.New("nil transformer from config: "+config))
				continue
			}
			tt.Response = append(tt.Response, transformer)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return tt, nil
}

func (tt TrafficTransformers) String() string {
	var requestStrings []string
	for _, transformer := range tt.Request {
		requestStrings = append(requestStrings, transformer.String())
	}

	var responseStrings []string
	for _, transformer := range tt.Response {
		responseStrings = append(responseStrings, transformer.String())
	}

	return fmt.Sprintf("Request Transformers:\n%s\nResponse Transformers:\n%s",
		strings.Join(requestStrings, "\n"), strings.Join(responseStrings, "\n"))
}
