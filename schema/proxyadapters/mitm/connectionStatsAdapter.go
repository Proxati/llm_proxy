package mitm

import px "github.com/proxati/mitmproxy/proxy"

const unknownAddr = "unknown"

// ConnectionStatsAdapter is an adapter for the connection stats object
type ConnectionStatsAdapter struct {
	f *px.Flow
}

// NewProxyConnectionStatsAdapter creates a new connection stats adapter object
func NewProxyConnectionStatsAdapter(f *px.Flow) *ConnectionStatsAdapter {
	if f == nil {
		return nil
	}

	return &ConnectionStatsAdapter{f: f}
}

// GetClientIP returns the client IP address, to implement the ConnectionStatsReaderAdapter interface
func (cs *ConnectionStatsAdapter) GetClientIP() string {
	if cs.f == nil || cs.f.ConnContext == nil || cs.f.ConnContext.ClientConn == nil || cs.f.ConnContext.ClientConn.Conn == nil {
		// Ugh != nil
		return unknownAddr
	}
	remote := cs.f.ConnContext.ClientConn.Conn.RemoteAddr()
	if remote == nil {
		return unknownAddr
	}
	return remote.String()
}

// GetProxyID returns the proxy ID, to implement the ConnectionStatsReaderAdapter interface
func (cs *ConnectionStatsAdapter) GetProxyID() string {
	if cs.f == nil {
		return ""
	}
	return cs.f.Id.String()
}

// GetRequestURL returns the request URL, to implement the ConnectionStatsReaderAdapter interface
func (cs *ConnectionStatsAdapter) GetRequestURL() string {
	if cs.f == nil || cs.f.Request == nil || cs.f.Request.URL == nil {
		return ""
	}
	return cs.f.Request.URL.String()
}
