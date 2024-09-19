package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterNestedInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		prefix   string
		options  map[string]string
		expected map[string]string
	}{
		{
			name:   "MatchingPrefix",
			prefix: "health-check",
			options: map[string]string{
				"health-check.interval": "30s",
				"health-check.path":     "/status",
				"timeout":               "10s",
			},
			expected: map[string]string{
				"interval": "30s",
				"path":     "/status",
			},
		},
		{
			name:   "NoMatchingPrefix",
			prefix: "health-check",
			options: map[string]string{
				"timeout":     "10s",
				"concurrency": "5",
			},
			expected: map[string]string{},
		},
		{
			name:   "MixedMatchingAndNonMatching",
			prefix: "health-check",
			options: map[string]string{
				"health-check.interval": "30s",
				"timeout":               "10s",
				"health-check.path":     "/status",
			},
			expected: map[string]string{
				"interval": "30s",
				"path":     "/status",
			},
		},
		{
			name:     "EmptyOptions",
			prefix:   "health-check",
			options:  map[string]string{},
			expected: map[string]string{},
		},
		{
			name:   "EmptyPrefix",
			prefix: "",
			options: map[string]string{
				"interval": "30s",
				"path":     "/status",
			},
			expected: map[string]string{},
		},
		{
			name:   "NestedPrefix",
			prefix: "health-check.sub",
			options: map[string]string{
				"health-check.sub.interval": "30s",
				"health-check.sub.path":     "/status",
				"timeout":                   "10s",
				"health-check.interval":     "30s",
			},
			expected: map[string]string{
				"interval": "30s",
				"path":     "/status",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterNestedInput(tt.prefix, tt.options)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseStructTags(t *testing.T) {
	t.Parallel()
	type SubStruct struct {
		SubField1 string        `transformer:"sub-field1"`
		SubField2 time.Duration `transformer:"sub-field2"`
	}

	type TestStruct struct {
		Field1    string        `transformer:"field1"`
		Field2    int           `transformer:"field2"`
		Field3    bool          `transformer:"field3"`
		Field4    float64       `transformer:"field4"`
		Field5    time.Duration `transformer:"field5"`
		SubStruct SubStruct     `transformer:"sub-struct"`
	}

	tests := []struct {
		name     string
		input    map[string]string
		expected TestStruct
		hasError bool
	}{
		{
			name: "ValidInput_AllFields",
			input: map[string]string{
				"field1":                "value1",
				"field2":                "42",
				"field3":                "true",
				"field4":                "3.14",
				"field5":                "2s",
				"sub-struct.sub-field1": "subvalue1",
				"sub-struct.sub-field2": "1m",
			},
			expected: TestStruct{
				Field1: "value1",
				Field2: 42,
				Field3: true,
				Field4: 3.14,
				Field5: 2 * time.Second,
				SubStruct: SubStruct{
					SubField1: "subvalue1",
					SubField2: 1 * time.Minute,
				},
			},
			hasError: false,
		},
		{
			name: "ValidInput_MinimumFields",
			input: map[string]string{
				"field1": "value1",
			},
			expected: TestStruct{
				Field1: "value1",
			},
			hasError: false,
		},
		{
			name: "InvalidInput_InvalidInt",
			input: map[string]string{
				"field2": "invalid",
			},
			expected: TestStruct{},
			hasError: true,
		},
		{
			name: "InvalidInput_InvalidBool",
			input: map[string]string{
				"field3": "invalid",
			},
			expected: TestStruct{},
			hasError: true,
		},
		{
			name: "InvalidInput_InvalidFloat",
			input: map[string]string{
				"field4": "invalid",
			},
			expected: TestStruct{},
			hasError: true,
		},
		{
			name: "InvalidInput_InvalidDuration",
			input: map[string]string{
				"field5": "invalid",
			},
			expected: TestStruct{},
			hasError: true,
		},
		{
			name: "InvalidInput_InvalidSubStructDuration",
			input: map[string]string{
				"sub-struct.sub-field2": "invalid",
			},
			expected: TestStruct{},
			hasError: true,
		},
		{
			name: "InvalidInput_UnknownField",
			input: map[string]string{
				"unknown": "value",
			},
			expected: TestStruct{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := parseStructTags("transformer", tt.input, reflect.ValueOf(&result))
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
