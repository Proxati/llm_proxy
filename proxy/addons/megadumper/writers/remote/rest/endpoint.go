package rest

// Endpoint represents a single REST endpoint
type Endpoint interface {
	String() string // name of the endpoint
	GetURL() string // URL of the endpoint
	POST(identifier string, data []byte) error
}
