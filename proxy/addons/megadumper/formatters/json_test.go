package formatters

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/proxati/llm_proxy/v2/schema"
	"github.com/stretchr/testify/assert"
)

func TestJSONFormatter(t *testing.T) {
	container := schema.NewLogDumpContainerWithDefaults()
	container.Request = &schema.ProxyRequest{
		Header: http.Header{"ReqHeader": []string{"ReqValue"}},
		Body:   "Request Body",
	}
	container.Response = &schema.ProxyResponse{
		Header: http.Header{"RespHeader": []string{"RespValue"}},
		Body:   "Response Body",
	}

	j := &JSON{}

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	// Create expected container and marshal to JSON
	expectedContainer := &schema.LogDumpContainer{
		ObjectType:    container.ObjectType,
		SchemaVersion: container.SchemaVersion,
		Timestamp:     container.Timestamp,
		Request: &schema.ProxyRequest{
			Header: http.Header{"ReqHeader": []string{"ReqValue"}},
			Body:   "Request Body",
		},
		Response: &schema.ProxyResponse{
			Header: http.Header{"RespHeader": []string{"RespValue"}},
			Body:   "Response Body",
		},
	}

	expectedJSONBytes, err := json.MarshalIndent(expectedContainer, "", "  ")
	assert.NoError(t, err)

	var expectedParsedJSON map[string]interface{}
	err = json.Unmarshal(expectedJSONBytes, &expectedParsedJSON)
	assert.NoError(t, err)

	assert.Equal(t, expectedParsedJSON, parsedJSON)
}

func TestJSONFormatter_Empty(t *testing.T) {
	container := schema.NewLogDumpContainerWithDefaults()
	container.Request = &schema.ProxyRequest{}
	container.Response = &schema.ProxyResponse{}

	j := &JSON{}

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	// Create expected container and marshal to JSON
	expectedContainer := &schema.LogDumpContainer{
		ObjectType:    container.ObjectType,
		SchemaVersion: container.SchemaVersion,
		Timestamp:     container.Timestamp,
		Request:       &schema.ProxyRequest{},
		Response:      &schema.ProxyResponse{},
	}

	expectedJSONBytes, err := json.MarshalIndent(expectedContainer, "", "  ")
	assert.NoError(t, err)

	var expectedParsedJSON map[string]interface{}
	err = json.Unmarshal(expectedJSONBytes, &expectedParsedJSON)
	assert.NoError(t, err)

	assert.Equal(t, expectedParsedJSON, parsedJSON)
}

func TestJSONFormatter_NilContainer(t *testing.T) {
	j := &JSON{}
	jsonBytes, err := j.Read(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("{}"), jsonBytes)
}

func TestJSONFormatter_implements_Reader(t *testing.T) {
	var _ MegaDumpFormatter = &JSON{}
}
