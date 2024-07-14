package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToTrafficLogFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected TrafficLogFormat
		err      bool
	}{
		{"json", TrafficLog_JSON, false},
		{"Json", TrafficLog_JSON, false},
		{"JSON", TrafficLog_JSON, false},
		{"txt", TrafficLog_TXT, false},
		{"Txt", TrafficLog_TXT, false},
		{"TXT", TrafficLog_TXT, false},
		{"text", TrafficLog_TXT, false},
		{"Text", TrafficLog_TXT, false},
		{"TEXT", TrafficLog_TXT, false},
		{"unsupported", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			format, err := StringToTrafficLogFormat(test.input)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, format)
			}
		})
	}
}
