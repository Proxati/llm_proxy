package openai

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEmbeddedDataJSON(t *testing.T) {
	assert.NotEmpty(t, APIEndpointData, "init() populates this variable")

	// Reset API_Endpoint_Pricing before test
	APIEndpointData = nil

	err := loadEmbeddedDataJSON()
	assert.Nil(t, err, "Expected no error loading data.json, but got an error")

	assert.NotEmpty(t, APIEndpointData, "Expected API_Endpoint_Pricing to be populated, but it was empty")

	// alphabetize the list of products, to confirm the JSON file is sorted correctly
	for _, endpoint := range APIEndpointData {
		unSortedProducts := make([]Product, len(endpoint.Products))
		copy(unSortedProducts, endpoint.Products)

		sort.Slice(endpoint.Products, func(i, j int) bool {
			return endpoint.Products[i].Name < endpoint.Products[j].Name
		})
		assert.Equal(t, unSortedProducts, endpoint.Products, fmt.Sprintf("Expected products to be sorted alphabetically for endpoint: %s", endpoint.URL))
	}
}
