package schema

import (
	"encoding/json"

	px "github.com/proxati/mitmproxy/proxy"
)

const UnknownAddr = "unknown"

type ConnectionStatsContainer struct {
	ClientAddress string `json:"client_address"`
	URL           string `json:"url"`
	Duration      int64  `json:"duration_ms"`
	ProxyID       string `json:"proxy_id,omitempty"`
}

func (obj *ConnectionStatsContainer) ToJSON() []byte {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		getLogger().Error("Could not convert object to JSON", "error", err)
		return []byte("{}")
	}
	return jsonData
}

func (obj *ConnectionStatsContainer) ToJSONstr() string {
	return string(obj.ToJSON())
}

func getClientAddr(f *px.Flow) string {
	if f == nil || f.ConnContext == nil || f.ConnContext.ClientConn == nil || f.ConnContext.ClientConn.Conn == nil {
		// Ugh != nil
		return UnknownAddr
	}
	remote := f.ConnContext.ClientConn.Conn.RemoteAddr()
	if remote == nil {
		return UnknownAddr
	}
	return remote.String()
}

func newConnectionStatusContainer(f *px.Flow) *ConnectionStatsContainer {
	if f == nil {
		return &ConnectionStatsContainer{}
	}
	logOutput := &ConnectionStatsContainer{
		ClientAddress: getClientAddr(f),
		ProxyID:       f.Id.String(),
	}
	if f.Request != nil && f.Request.URL != nil {
		logOutput.URL = f.Request.URL.String()
	}
	return logOutput
}

// NewConnectionStatusContainerWithDuration is a slightly leaky abstraction, the doneAt param is for logging
// the entire session length, and comes from the proxy addon layer.
func NewConnectionStatusContainerWithDuration(f *px.Flow, doneAt int64) *ConnectionStatsContainer {
	logOutput := newConnectionStatusContainer(f)
	logOutput.Duration = doneAt
	return logOutput
}
