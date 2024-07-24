package writers_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"testing"

	"github.com/proxati/llm_proxy/v2/proxy/addons/megadumper/writers"
	"github.com/stretchr/testify/assert"
)

type testLogWriter struct {
	buf bytes.Buffer
}

func (w *testLogWriter) Close() error {
	return nil
}

func (w *testLogWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

func TestToStdOut_Write(t *testing.T) {
	// Redirect log output to a buffer
	testWriter := &testLogWriter{}
	logger := slog.New(slog.NewTextHandler(testWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Create a new ToStdOut instance
	toStdOut, err := writers.NewToStdOut()
	assert.NoError(t, err)

	// Write some data
	identifier := "test"
	data := []byte("test data")
	bytesWritten, err := toStdOut.Write(identifier, data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), bytesWritten)

	// Assert on the captured output
	assert.Contains(t, testWriter.buf.String(), fmt.Sprintf(`identifier=%s`, identifier))
	assert.Contains(t, testWriter.buf.String(), fmt.Sprintf(`msg="%s"`, string(data)))
}
