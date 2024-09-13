package headers

const (
	// ProxyID is a response header with a request ID assigned by this proxy to track the request
	ProxyID = "X-Llm_proxy-id"

	// ProxyVersion is a response header that indicates the version of this proxy
	Version = "X-Llm_proxy-version"

	// SchemeUpgraded is a response header that indicates that the scheme was upgraded (http->https)
	SchemeUpgraded = "X-Llm_proxy-scheme-upgraded"

	// WorkflowName is an optional request header that can be used to specify the name of the workflow
	WorkflowName = "X-Llm_workflow-name"

	// CacheStatusHeader is a response header that indicates the cache status of the response
	CacheStatusHeader    = "X-Llm_proxy-Cache"
	CacheStatusValueHit  = "HIT"
	CacheStatusValueMiss = "MISS"
	CacheStatusValueSkip = "SKIP"
)
