package schema

import (
	"testing"

	"github.com/bojanz/currency"
	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPI_Provider(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		provider, err := newAPIProvider("test", "model", "0.01", "0.02", "USD")
		assert.NotNil(t, provider)
		assert.NoError(t, err)
	})
	t.Run("Invalid input cost", func(t *testing.T) {
		provider, err := newAPIProvider("test", "model", "invalid", "0.02", "USD")
		assert.Nil(t, provider)
		assert.Error(t, err)
	})

	t.Run("Invalid output cost", func(t *testing.T) {
		provider, err := newAPIProvider("test", "model", "0.01", "invalid", "USD")
		assert.Nil(t, provider)
		assert.Error(t, err)
	})

	t.Run("Invalid currency type", func(t *testing.T) {
		provider, err := newAPIProvider("test", "model", "0.01", "0.02", "nope")
		assert.Nil(t, provider)
		assert.Error(t, err)
	})
}

func TestAPI_ProviderString(t *testing.T) {
	provider, _ := newAPIProvider("test", "model", "0.01", "0.02", "USD")
	provider.totalCost, _ = currency.NewAmount("1", "USD")
	assert.Equal(t, "1.00 USD", provider.String())
}

func TestAddRequest(t *testing.T) {
	provider, _ := newAPIProvider("test", "model", "0.01", "0.02", "USD")
	req := &ProxyRequest{}
	chatCompReq := &openai.ChatCompletionRequest{}
	provider.addRequest(req, chatCompReq)
	assert.Equal(t, 1, len(provider.apiRequests))
	assert.Equal(t, 1, len(provider.apiRequestBodies))
}

func TestAddResponse(t *testing.T) {
	t.Run("Normal cost summing", func(t *testing.T) {
		provider, _ := newAPIProvider("test", "model", "0.01", "0.02", "USD")
		resp := &ProxyResponse{}
		chatCompResp := &openai.ChatCompletionResponse{
			Usage: openai.Usage{
				PromptTokens:     10,
				CompletionTokens: 10,
			},
		}
		provider.addResponse(resp, chatCompResp)
		assert.Equal(t, 1, len(provider.apiResponses))
		assert.Equal(t, 1, len(provider.apiResponseBodies))

		require.Equal(t, "0 USD", provider.totalCost.String()) // not calculated yet
		provider.calculateCost(chatCompResp)

		// check the cost: 0.01 * 10 + 0.02 * 10 = 0.30
		expectedCost, _ := currency.NewAmount("0.30", "USD")
		require.Equal(t, expectedCost.String(), provider.totalCost.String())

		// add another response
		provider.addResponse(resp, chatCompResp)
		provider.calculateCost(chatCompResp)
		assert.Equal(t, 2, len(provider.apiResponses))
		assert.Equal(t, 2, len(provider.apiResponseBodies))

		// check the cost: 0.30 + 0.30 = 0.60
		expectedCost, _ = currency.NewAmount("0.60", "USD")
		require.Equal(t, expectedCost.String(), provider.totalCost.String())

		// add another response with different token counts
		chatCompResp.Usage.PromptTokens = 20
		chatCompResp.Usage.CompletionTokens = 200

		provider.addResponse(resp, chatCompResp)
		provider.calculateCost(chatCompResp)
		assert.Equal(t, 3, len(provider.apiResponses))
		assert.Equal(t, 3, len(provider.apiResponseBodies))

		// check the cost: 0.60 + 0.20 + 4.00 = 4.80
		expectedCost, _ = currency.NewAmount("4.80", "USD")
		require.Equal(t, expectedCost.String(), provider.totalCost.String())
	})

	t.Run("Empty cost summing", func(t *testing.T) {
		provider, _ := newAPIProvider("test", "model", "0.01", "0.02", "USD")
		resp := &ProxyResponse{}
		chatCompResp := &openai.ChatCompletionResponse{} // empty usage, no tokens spent

		provider.addResponse(resp, chatCompResp)
		assert.Equal(t, 1, len(provider.apiResponses))
		assert.Equal(t, 1, len(provider.apiResponseBodies))

		require.Equal(t, "0 USD", provider.totalCost.String()) // not calculated yet
		provider.calculateCost(chatCompResp)
		require.Equal(t, "0.00 USD", provider.totalCost.String()) // formatted as 0.00 USD after being calculated
	})
}
