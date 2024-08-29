package openai

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAICompletionResponse(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectError    bool
		expectedResult *openai.ChatCompletionRequest
	}{
		{
			name: "Valid JSON",
			body: `
{
	"messages": [{
		"role": "user", "content": "Hello, you are amazing."
	}],
	"model": "gpt-3.5-turbo"
}`,
			expectError: false,
			expectedResult: &openai.ChatCompletionRequest{
				Model: "gpt-3.5-turbo",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    "user",
						Content: "Hello, you are amazing.",
					},
				},
			},
		},
		{
			name:           "Invalid JSON",
			body:           `{"choices": [}`,
			expectError:    true,
			expectedResult: nil,
		},
		{
			name:           "Valid JSON, wrong structure",
			body:           `{"foo": "bar"}`,
			expectError:    false,
			expectedResult: &openai.ChatCompletionRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &tt.body
			result, err := NewOpenAIChatCompletionRequest(body)

			if tt.expectError {
				assert.Error(t, err)
				require.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestNewOpenAIChatCompletionResponse(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectError    bool
		expectedResult *openai.ChatCompletionResponse
	}{
		{
			name: "Valid JSON",
			body: `{
	"id": "chatcmpl-9aCnlK7qGO03XZTGVp8NjC1gyxMSk",
	"object": "chat.completion",
	"created": 1718416045,
	"model": "gpt-3.5-turbo-0125",
	"choices": [{
		"index": 0,
		"message": {
			"role": "assistant",
			"content": "Thank you"
		},
		"logprobs": null,
		"finish_reason": "stop"
	}],
	"usage": {
		"prompt_tokens": 13,
		"completion_tokens": 20,
		"total_tokens": 33
	},
	"system_fingerprint": null
}`,
			expectError: false,
			expectedResult: &openai.ChatCompletionResponse{
				ID:      "chatcmpl-9aCnlK7qGO03XZTGVp8NjC1gyxMSk",
				Object:  "chat.completion",
				Created: 1718416045,
				Model:   "gpt-3.5-turbo-0125",
				Usage: openai.Usage{
					PromptTokens:     13,
					CompletionTokens: 20,
					TotalTokens:      33,
				},
				Choices: []openai.ChatCompletionChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionMessage{
							Role:    "assistant",
							Content: "Thank you",
						},
						LogProbs:     nil,
						FinishReason: "stop",
					},
				},
			},
		},
		{
			name:           "Invalid JSON",
			body:           `{"id": "cmpl-2C4k9l8A8K9Z8F2eVJGxoXfg", "choices": [}`,
			expectError:    true,
			expectedResult: nil,
		},
		{
			name:           "Valid JSON, missing fields",
			body:           `{"foo": "bar"}`,
			expectError:    false,
			expectedResult: &openai.ChatCompletionResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &tt.body
			result, err := NewOpenAIChatCompletionResponse(body)
			if tt.expectError {
				assert.Error(t, err)
				require.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
