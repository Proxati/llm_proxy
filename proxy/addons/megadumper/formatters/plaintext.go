package formatters

import (
	"bytes"

	"github.com/proxati/llm_proxy/schema"
)

const text_ext = "log"

type PlainText struct {
	container *schema.LogDumpContainer
}

func (pt *PlainText) flatten() ([]byte, error) {
	if pt.container == nil {
		return []byte(""), nil
	}

	buf := new(bytes.Buffer)

	if pt.container.Request.Header != nil {
		buf.WriteString(pt.container.Request.HeaderString())
	}

	if pt.container.Request.Body != "" {
		buf.WriteString(pt.container.Request.Body)
		buf.WriteString("\r\n")
	}

	if pt.container.Response.Header != nil {
		buf.WriteString(pt.container.Response.HeaderString())
	}

	if pt.container.Response.Body != "" {
		buf.WriteString(pt.container.Response.Body)
		buf.WriteString("\r\n")
	}

	return buf.Bytes(), nil

}

// Read returns a flattened representation of all the fields in the LogDumpContainer
func (pt *PlainText) Read(container *schema.LogDumpContainer) ([]byte, error) {
	pt.container = container
	return pt.flatten()
}

// GetFileExtension returns the file extension for a plain text file
func (pt *PlainText) GetFileExtension() string {
	return text_ext
}
