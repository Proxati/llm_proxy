package schema

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/bojanz/currency"
	openai "github.com/proxati/llm_proxy/v2/schema/providers/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCostCounter(t *testing.T) {
	cc := NewCostCounterDefaults()
	assert.NotNil(t, cc)
	assert.NotNil(t, cc.providers)
	assert.NotNil(t, cc.lookupCache)
}

func TestCostCounterString(t *testing.T) {
	mon, err := currency.NewAmount("100", "USD")
	assert.NoError(t, err)

	cc := &CostCounter{
		grandTotal: mon,
	}

	assert.Equal(t, "100 USD", cc.String())
}

func TestProviderLookup(t *testing.T) {
	cc := NewCostCounterDefaults()
	// Assuming there's a provider with URL "http://example.com" and product "testModel" for testing
	productURL := openai.APIEndpointData[0]
	model := productURL.Products[0]

	provider := cc.providerLookup(productURL.URL, model.Name)
	assert.NotNil(t, provider)

	// Verify cache functionality
	cachedProvider := cc.providerLookup(productURL.URL, model.Name)
	assert.Equal(t, provider, cachedProvider)

	// Invalid provider
	invalidProvider := cc.providerLookup("invalid_url", "invalid_product")
	assert.Nil(t, invalidProvider)

	// Verify that the cache is used
	cacheKey := productURL.URL + "|" + model.Name
	cachedProvider = cc.providerCacheLookup(cacheKey)
	assert.NotNil(t, cachedProvider)

	// Verify that the cache is not used when the key is not found
	cachedProvider = cc.providerCacheLookup("invalid_key")
	assert.Nil(t, cachedProvider)
}

func TestAdd(t *testing.T) {
	cc := NewCostCounterDefaults()
	// Setup: Assuming there's a provider with URL "http://example.com" and product "testModel" for testing
	productURL := openai.APIEndpointData[0]
	model := productURL.Products[0]
	provider := cc.providerLookup(productURL.URL, model.Name)

	reqURL, _ := url.Parse(productURL.URL)
	req := ProxyRequest{
		URL:  reqURL,
		Body: `{"model": "testModel", "prompt": "Hello, world!"}`,
	}
	resp := ProxyResponse{
		Body: `{"choices": [{"text": "Hello!"}]}`,
	}
	expectedOutput := &AuditOutput{
		URL:          productURL.URL,
		Model:        model.Name,
		InputCost:    "$0.00",
		OutputCost:   "$0.00",
		TotalReqCost: "$0.00",
		GrandTotal:   "$0.00",
	}

	out, err := cc.Add(req, resp)
	require.Error(t, err, "Error when invalid model is used in request")
	require.Nil(t, out)
	assert.Len(t, provider.apiRequests, 0)
	assert.Len(t, provider.apiRequestBodies, 0)
	assert.Len(t, provider.apiResponses, 0)
	assert.Len(t, provider.apiResponseBodies, 0)

	// Valid request
	req = ProxyRequest{
		URL:  reqURL,
		Body: fmt.Sprintf(`{"model": "%s", "prompt": "Hello, world!"}`, model.Name),
	}
	out, err = cc.Add(req, resp)
	require.NoError(t, err)
	require.Equal(t, expectedOutput, out)
	assert.Len(t, provider.apiRequests, 1)
	assert.Len(t, provider.apiRequestBodies, 1)
	assert.Len(t, provider.apiResponses, 1)
	assert.Len(t, provider.apiResponseBodies, 1)
}
