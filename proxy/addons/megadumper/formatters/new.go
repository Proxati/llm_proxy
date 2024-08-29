package formatters

import (
	"fmt"

	"github.com/proxati/llm_proxy/v2/config"
)

// NewMegaDumpFormatter returns an object that implements the MegaDumpFormatter interface based on
// the requested format.
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
