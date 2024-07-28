package mitm

import px "github.com/proxati/mitmproxy/proxy"

const UnknownAddr = "unknown"

type ConnectionStatsAdapter struct {
	f *px.Flow
}

func NewProxyConnectionStatsAdapter(f *px.Flow) *ConnectionStatsAdapter {
	if f == nil {
		return nil
	}

	return &ConnectionStatsAdapter{f: f}
}

func (cs *ConnectionStatsAdapter) GetClientIP() string {
	if cs.f == nil || cs.f.ConnContext == nil || cs.f.ConnContext.ClientConn == nil || cs.f.ConnContext.ClientConn.Conn == nil {
		// Ugh != nil
		return UnknownAddr
	}
	remote := cs.f.ConnContext.ClientConn.Conn.RemoteAddr()
	if remote == nil {
		return UnknownAddr
	}
	return remote.String()
}

func (cs *ConnectionStatsAdapter) GetProxyID() string {
	if cs.f == nil {
		return ""
	}
	return cs.f.Id.String()
}

func (cs *ConnectionStatsAdapter) GetRequestURL() string {
	if cs.f == nil || cs.f.Request == nil || cs.f.Request.URL == nil {
		return ""
	}
	return cs.f.Request.URL.String()
}
