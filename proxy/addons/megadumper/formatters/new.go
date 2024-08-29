package formatters

import (
	"fmt"

	"github.com/proxati/llm_proxy/v2/config"
)

// MegaDumpFormatter is an interface for formatting log dumps
func NewMegaDumpFormatter(format config.LogFormat) (MegaDumpFormatter, error) {
	var f MegaDumpFormatter

	switch format {
	case config.LogFormatJSON:
		f = &JSON{}
	case config.LogFormatTXT:
		f = &PlainText{}
	default:
		return nil, fmt.Errorf("unsupported log format: %v", format)
	}

	return f, nil
}
