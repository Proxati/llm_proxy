package openai

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
)

//go:embed data.json
var pricingDataJSON embed.FS

// APIEndpointData is populated from init() with data loaded from the embedded JSON file
var APIEndpointData []APIEndpoint

// Product represents a model or other product attached to an endpoint
type Product struct {
	Name            string `json:"name"`
	InputTokenCost  string `json:"inputTokenCost"`
	OutputTokenCost string `json:"outputTokenCost"`
	Currency        string `json:"currency"`
}

// APIEndpoint represents the pricing data for a single API endpoint, such as "https://api.openai.com/v1/chat/completions"
type APIEndpoint struct {
	URL      string    `json:"url"`
	Products []Product `json:"products"`
}

func loadEmbeddedDataJSON() error {
	data, err := fs.ReadFile(pricingDataJSON, "data.json")
	if err != nil {
		return fmt.Errorf("failed to read embedded data.json: %w", err)
	}
	return json.Unmarshal(data, &APIEndpointData)
}

func init() {
	err := loadEmbeddedDataJSON()
	if err != nil {
		panic(fmt.Sprintf("Error loading openai pricing data: %v\n", err))
	}
}
