package config

type AppMode int

const (
	ProxyRunMode AppMode = iota
	CacheMode
	APIAuditMode
)

func (a AppMode) String() string {
	switch a {
	case ProxyRunMode:
		return "ProxyRunMode"
	case CacheMode:
		return "CacheMode"
	case APIAuditMode:
		return "APIAuditMode"
	default:
		return "Unknown"
	}
}
