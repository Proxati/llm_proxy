package config

type AppMode int

const (
	SimpleMode AppMode = iota
	CacheMode
	APIAuditMode
)
