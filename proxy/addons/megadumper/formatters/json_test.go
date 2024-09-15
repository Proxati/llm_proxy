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

	// Format the container.Timestamp back into a string
	const layout = "2006-01-02T15:04:05Z07:00"
	timestampStr := container.Timestamp.Format(layout)

	j := &JSON{}

	expectedJSON := `{
		"timestamp": "` + timestampStr + `",
		"request": {
		  "body": "Request Body",
		  "header": { "ReqHeader": [ "ReqValue" ] }
		},
		"response": {
		  "body": "Response Body",
		  "header": { "RespHeader": [ "RespValue" ] }
		}
	  }`

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	keys := []string{"timestamp", "request", "response"}
	for _, key := range keys {
		parsedValue, ok := parsedJSON[key]
		if ok {
			expectedValue, ok := expectedParsedJSON[key]
			if ok {
				assert.Equal(t, expectedValue, parsedValue)
			} else {
				t.Errorf("Expected to find %s in expectedParsedJSON", key)
			}
		} else {
			t.Errorf("Expected to find %s in parsedJSON", key)
		}
	}
}

func TestJSONFormatter_Empty(t *testing.T) {
	container := schema.NewLogDumpContainerWithDefaults()
	container.Request = &schema.ProxyRequest{}
	container.Response = &schema.ProxyResponse{}
	// Format the container.Timestamp back into a string
	const layout = "2006-01-02T15:04:05Z07:00"
	timestampStr := container.Timestamp.Format(layout)

	j := &JSON{}
	expectedJSON := `{
		"timestamp": "` + timestampStr + `",
		"request": {
			"body": "",
			"header": null
		},
		"response": {
			"body": "",
			"header": null
		}
	  }`

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	keys := []string{"timestamp", "request", "response"}
	for _, key := range keys {
		parsedValue, ok := parsedJSON[key]
		if ok {
			expectedValue, ok := expectedParsedJSON[key]
			if ok {
				assert.Equal(t, expectedValue, parsedValue)
			} else {
				t.Errorf("Expected to find %s in expectedParsedJSON", key)
			}
		} else {
			t.Errorf("Expected to find %s in parsedJSON", key)
		}
	}

}

func TestJSONFormatter_implements_Reader(t *testing.T) {
	var _ MegaDumpFormatter = &JSON{}
}
