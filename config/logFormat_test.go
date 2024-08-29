package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToTrafficLogFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input          string
		expected       LogFormat
		expectedString string
		err            bool
	}{
		{"json", LogFormatJSON, "json", false},
		{"Json", LogFormatJSON, "json", false},
		{"JSON", LogFormatJSON, "json", false},
		{"JSON ", LogFormatJSON, "json", false},
		{" JSON ", LogFormatJSON, "json", false},
		{" JSON", LogFormatJSON, "json", false},
		{"txt", LogFormatTXT, "txt", false},
		{"Txt", LogFormatTXT, "txt", false},
		{"TXT", LogFormatTXT, "txt", false},
		{"text", LogFormatTXT, "txt", false},
		{"Text", LogFormatTXT, "txt", false},
		{"TEXT", LogFormatTXT, "txt", false},
		{"TEXT ", LogFormatTXT, "txt", false},
		{" TEXT ", LogFormatTXT, "txt", false},
		{" TEXT", LogFormatTXT, "txt", false},
		{"unsupported", 0, "", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			format, err := StringToLogFormat(test.input)
			if test.err {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expected, format)
			assert.Equal(t, test.expectedString, format.String())
		})
	}
}
