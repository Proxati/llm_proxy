package schema

import px "github.com/proxati/mitmproxy/proxy"

type ConnectionStatsReaderAdapter interface {
	GetClientIP() string
	GetProxyID() string
	GetRequestURL() string
}

type ConnectionStatsAdapter_MiTM struct {
	f *px.Flow
}

func newProxyConnectionStatsAdapter_MiTM(f *px.Flow) *ConnectionStatsAdapter_MiTM {
	if f == nil {
		return nil
	}

	return &ConnectionStatsAdapter_MiTM{f: f}
}

func (cs *ConnectionStatsAdapter_MiTM) GetClientIP() string {
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

func (cs *ConnectionStatsAdapter_MiTM) GetProxyID() string {
	if cs.f == nil {
		return ""
	}
	return cs.f.Id.String()
}

func (cs *ConnectionStatsAdapter_MiTM) GetRequestURL() string {
	if cs.f == nil || cs.f.Request == nil || cs.f.Request.URL == nil {
		return ""
	}
	return cs.f.Request.URL.String()
}
