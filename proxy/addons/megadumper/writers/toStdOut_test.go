package writers_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/proxati/llm_proxy/proxy/addons/megadumper/writers"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestToStdOut_Write(t *testing.T) {
	// Redirect standard output to a buffer
	origStdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	log.SetOutput(w)

	// Capture output
	var buf bytes.Buffer
	outputCapture := make(chan string)
	go func() {
		io.Copy(&buf, r)
		outputCapture <- buf.String()
	}()

	// Create a new ToStdOut instance
	toStdOut, err := writers.NewToStdOut()
	assert.NoError(t, err)

	// Write some data
	identifier := "test"
	data := []byte("test data")
	_, err = toStdOut.Write(identifier, data)
	assert.NoError(t, err)

	// Close the write end of the pipe to finish the capture
	w.Close()
	os.Stdout = origStdOut // Restore original standard output
	log.SetOutput(os.Stdout)

	// Assert on the captured output
	capturedOutput := <-outputCapture
	assert.Contains(t, capturedOutput, identifier)
	assert.Contains(t, capturedOutput, string(data))
}
