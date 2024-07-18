package cache

import (
	"github.com/proxati/llm_proxy/v2/schema"
)

type DB interface {
	Close() error
	Len(identifier string) (int, error)
	Get(identifier string, body []byte) (response *schema.ProxyResponse, err error)
	Put(request *schema.ProxyRequest, response *schema.ProxyResponse) error
}
