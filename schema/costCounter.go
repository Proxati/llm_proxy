package schema

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bojanz/currency"
	openai "github.com/proxati/llm_proxy/v2/schema/providers/openai"
)

const (
	OutputFormatFull    = "URL: {url} Model: {model} inputCost: {inputCost} outputCost {outputCost} = Request Cost: {totalReqCost} Grand Total: {grandTotal}"
	OutputFormatCompact = "Request Cost: {totalReqCost} Grand Total: {grandTotal}"
)

// AuditOutput is a struct that holds the output data (cost totals) from a single transaction
type AuditOutput struct {
	URL          string `JSON:"url"`
	Model        string `JSON:"model"`
	InputCost    string `JSON:"inputCost"`
	OutputCost   string `JSON:"outputCost"`
	TotalReqCost string `JSON:"totalReqCost"`
	GrandTotal   string `JSON:"grandTotal"`
}

func (output *AuditOutput) String() string {
	return output.OutputStringFormatter(OutputFormatFull)
}

// OutputStringFormatter takes an AuditOutput and a format string and returns a formatted string
func (output *AuditOutput) OutputStringFormatter(formatString string) string {
	if formatString == "" {
		formatString = OutputFormatFull
	}

	data := map[string]string{
		"{url}":          output.URL,
		"{model}":        output.Model,
		"{inputCost}":    output.InputCost,
		"{outputCost}":   output.OutputCost,
		"{totalReqCost}": output.TotalReqCost,
		"{grandTotal}":   output.GrandTotal,
	}

	for key, value := range data {
		if value == "" {
			continue
		}
		formatString = strings.Replace(formatString, key, value, -1)
	}
	return formatString
}

// CostCounter is a struct that holds the state of the cost counter
type CostCounter struct {
	grandTotal  currency.Amount
	providers   map[string][]*API_Provider // key: provider URL, value: slice of models/products
	lookupCache map[string]*API_Provider   // key: provider URL + model, value: API_Provider
	formatter   *currency.Formatter
	rwMutex     sync.RWMutex
}

// NewCostCounter creates an object that _should_ be a singleton, in the Addon layer.
// currencyLocale is the locale for the currency formatter, e.g., "en-US"
// formatString is the output format string, e.g., OutputFormatFull or OutputFormatShort
func NewCostCounter(currencyLocale string) *CostCounter {
	loc := currency.NewLocale(currencyLocale) // "en-US" is the default

	cc := &CostCounter{
		providers:   make(map[string][]*API_Provider),
		lookupCache: make(map[string]*API_Provider),
		formatter:   currency.NewFormatter(loc),
		rwMutex:     sync.RWMutex{},
	}

	// iterate over the pricing data and populate this struct, data loaded from json in the openai package
	for _, provider := range openai.API_Endpoint_Data {
		for _, product := range provider.Products {
			apiProvider, err := newAPI_Provider(provider.URL, product.Name, product.InputTokenCost, product.OutputTokenCost, "USD")
			if err != nil {
				panic(fmt.Sprintf("Error creating API_Provider: %v", err))
			}
			cc.providers[provider.URL] = append(cc.providers[provider.URL], apiProvider)
		}
	}
	return cc
}

// NewCostCounterDefaults creates a new CostCounter object with reasonable defaults
func NewCostCounterDefaults() *CostCounter {
	return NewCostCounter("en-US")
}

func (cc *CostCounter) String() string {
	cc.rwMutex.RLock()
	defer cc.rwMutex.RUnlock()
	return cc.grandTotal.String()
}

func (cc *CostCounter) providerCacheLookup(cacheKey string) *API_Provider {
	cc.rwMutex.RLock()
	defer cc.rwMutex.RUnlock()

	if provider, found := cc.lookupCache[cacheKey]; found {
		return provider // Return the cached result
	}
	return nil
}

func (cc *CostCounter) providerLookup(url, product string) *API_Provider {
	cacheKey := fmt.Sprintf("%s|%s", url, product)

	// Check if the result is already in the cache
	if provider := cc.providerCacheLookup(cacheKey); provider != nil {
		// found it in the cache, return it
		return provider
	}

	providers := cc.providers[url]
	for _, provider := range providers {
		if provider.model == product {
			cc.rwMutex.Lock()
			defer cc.rwMutex.Unlock()

			cc.lookupCache[cacheKey] = provider // Cache the result
			return provider
		}
	}
	return nil
}

// Add is the primary method for working with this object, it takes a proxy req/resp
// and calculates the cost of the transaction, returning a struct with the output data
func (cc *CostCounter) Add(req ProxyRequest, resp ProxyResponse) (*AuditOutput, error) {
	// parse the request
	chatCompReq, err := openai.NewOpenAIChatCompletionRequest(&req.Body)
	if err != nil || chatCompReq == nil {
		return nil, fmt.Errorf("failed to create OpenAI completion request: %v", err)
	}

	// find the provider, which is a combination of the URL and the requested model
	provider := cc.providerLookup(req.URL.String(), chatCompReq.Model)
	if provider == nil {
		return nil, fmt.Errorf("provider not found for: %s|%s", req.URL.String(), chatCompReq.Model)
	}

	// store the request objects
	provider.addRequest(&req, chatCompReq)

	// parse the response object
	chatCompResp, err := openai.NewOpenAIChatCompletionResponse(&resp.Body)
	if err != nil || chatCompResp == nil {
		return nil, fmt.Errorf("failed to create OpenAI completion response: %v", err)
	}

	// store the response objects
	provider.addResponse(&resp, chatCompResp)

	// calculate the cost for this transaction
	inputCost, outputCost, err := provider.calculateCost(chatCompResp)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cost: %v", err)
	}

	// lock and update the grand total from the input/output token costs
	cc.rwMutex.Lock()
	defer cc.rwMutex.Unlock()

	cc.grandTotal, err = cc.grandTotal.Add(inputCost)
	if err != nil {
		return nil, fmt.Errorf("failed to add inputCost to the grand total: %v", err)
	}

	cc.grandTotal, err = cc.grandTotal.Add(outputCost)
	if err != nil {
		return nil, fmt.Errorf("failed to add outputCost to the grand total: %v", err)
	}
	totalReqCost, _ := inputCost.Add(outputCost)

	// return the output object with the formatted cost data w/ currency symbol added
	return &AuditOutput{
		URL:          req.URL.String(),
		Model:        chatCompReq.Model,
		InputCost:    cc.formatter.Format(inputCost),
		OutputCost:   cc.formatter.Format(outputCost),
		TotalReqCost: cc.formatter.Format(totalReqCost),
		GrandTotal:   cc.formatter.Format(cc.grandTotal),
	}, nil
}
