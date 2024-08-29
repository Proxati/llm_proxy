package config

// AppMode is an enum that represents different user-specified modes for the app
type AppMode int

const (
	// ProxyRunMode is the default mode, which runs the proxy without any additional features
	ProxyRunMode AppMode = iota

	// CacheMode runs the proxy with the literal cache feature enabled
	CacheMode

	// APIAuditMode runs the proxy with the API audit feature enabled, which shows the real-time cost for each API call
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
