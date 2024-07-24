package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToTrafficLogFormat(t *testing.T) {
	tests := []struct {
		input          string
		expected       LogFormat
		expectedString string
		err            bool
	}{
		{"json", LogFormat_JSON, "json", false},
		{"Json", LogFormat_JSON, "json", false},
		{"JSON", LogFormat_JSON, "json", false},
		{"JSON ", LogFormat_JSON, "json", false},
		{" JSON ", LogFormat_JSON, "json", false},
		{" JSON", LogFormat_JSON, "json", false},
		{"txt", LogFormat_TXT, "txt", false},
		{"Txt", LogFormat_TXT, "txt", false},
		{"TXT", LogFormat_TXT, "txt", false},
		{"text", LogFormat_TXT, "txt", false},
		{"Text", LogFormat_TXT, "txt", false},
		{"TEXT", LogFormat_TXT, "txt", false},
		{"TEXT ", LogFormat_TXT, "txt", false},
		{" TEXT ", LogFormat_TXT, "txt", false},
		{" TEXT", LogFormat_TXT, "txt", false},
		{"unsupported", 0, "", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			format, err := StringToLogFormat(test.input)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, format)
				assert.Equal(t, test.expectedString, format.String())
			}
		})
	}
}
