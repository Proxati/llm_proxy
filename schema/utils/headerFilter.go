package utils

type headerIndex map[string]any

func buildHeaderIndex(headers []string) headerIndex {
	index := make(headerIndex)
	for _, header := range headers {
		index[header] = nil
	}
	return index
}

type HeaderFilter struct {
	// Headers to be filtered from requests coming from the proxy
	Headers     []string
	headerIndex headerIndex
}

func NewHeaderFilter(headers []string) *HeaderFilter {
	return &HeaderFilter{
		Headers:     headers,
		headerIndex: buildHeaderIndex(headers),
	}
}

func (c *HeaderFilter) IsHeaderInIndex(header string) bool {
	_, exists := c.headerIndex[header]
	return exists
}
