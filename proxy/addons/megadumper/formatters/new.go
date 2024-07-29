package formatters

import (
	"fmt"

	"github.com/proxati/llm_proxy/v2/config"
)

// MegaDumpFormatter is an interface for formatting log dumps
func NewMegaDumpFormatter(format config.LogFormat) (MegaDumpFormatter, error) {
	var f MegaDumpFormatter

	switch format {
	case config.LogFormat_JSON:
		f = &JSON{}
	case config.LogFormat_TXT:
		f = &PlainText{}
	default:
		return nil, fmt.Errorf("unsupported log format: %v", format)
	}

	return f, nil
}
