package formatters

import (
	"fmt"
	"testing"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewMegaDumpFormatter(t *testing.T) {
	tests := []struct {
		format       config.LogFormat
		expectedType interface{}
		expectError  bool
	}{
		{config.LogFormatJSON, &JSON{}, false},
		{config.LogFormatTXT, &PlainText{}, false},
		{config.LogFormat(999), nil, true}, // Unsupported format
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("format=%v", tt.format), func(t *testing.T) {
			formatter, err := NewMegaDumpFormatter(tt.format)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, formatter)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.expectedType, formatter)
			}
		})
	}
}
