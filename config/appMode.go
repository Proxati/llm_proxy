package config

type AppMode int

const (
	ProxyRunMode AppMode = iota
	CacheMode
	APIAuditMode
)
