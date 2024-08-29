package openai

import (
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// NewOpenAIChatCompletionRequest creates a new OpenAI ChatCompletion Request object from a JSON string
func NewOpenAIChatCompletionRequest(body *string) (completion *openai.ChatCompletionRequest, err error) {
	bodyBytes := []byte(*body)

	err = json.Unmarshal(bodyBytes, &completion)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal OpenAI completion response body: %v", err)
	}

	return completion, nil
}

// NewOpenAIChatCompletionResponse creates a new OpenAI ChatCompletion Response object from a JSON string
func NewOpenAIChatCompletionResponse(body *string) (completion *openai.ChatCompletionResponse, err error) {
	bodyBytes := []byte(*body)

	err = json.Unmarshal(bodyBytes, &completion)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal OpenAI completion response body: %v", err)
	}

	return completion, nil
}
