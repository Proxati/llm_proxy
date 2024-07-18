package formatters

import (
	"net/http"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/stretchr/testify/assert"
)

func TestPlainText_flatten(t *testing.T) {
	container := &schema.LogDumpContainer{
		Request: &schema.ProxyRequest{
			Header: http.Header{"ReqHeader": []string{"ReqValue"}},
			Body:   "Request Body",
		},
		Response: &schema.ProxyResponse{
			Header: http.Header{"RespHeader": []string{"RespValue"}},
			Body:   "Response Body",
		},
	}
	pt := &PlainText{container}

	expectedResult := []byte("ReqHeader: ReqValue\r\nRequest Body\r\nRespHeader: RespValue\r\nResponse Body\r\n")

	result, err := pt.flatten()
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestPlainText_flatten_EmptyFields(t *testing.T) {
	pt := &PlainText{}

	expectedResult := []byte("")

	result, err := pt.flatten()
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
