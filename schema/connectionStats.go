package schema

import (
	"encoding/json"
)

const UnknownAddr = "unknown"

type ConnectionStats struct {
	ClientAddress string `json:"client_address"`
	URL           string `json:"url"`
	Duration      int64  `json:"duration_ms"`
	ProxyID       string `json:"proxy_id,omitempty"`
}

func (obj *ConnectionStats) ToJSON() []byte {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		getLogger().Error("Could not convert object to JSON", "error", err)
		return []byte("{}")
	}
	return jsonData
}

func (obj *ConnectionStats) ToJSONstr() string {
	return string(obj.ToJSON())
}

func newConnectionStats(cs ConnectionStatsReaderAdapter) *ConnectionStats {
	logOutput := &ConnectionStats{
		// ClientAddress: getClientAddr(f),
		// ProxyID:       f.Id.String(),
		ClientAddress: cs.GetClientIP(),
		ProxyID:       cs.GetProxyID(),
		URL:           cs.GetRequestURL(),
	}

	return logOutput
}

// NewConnectionStatsWithDuration is a slightly leaky abstraction, the doneAt param is for logging
// the entire session length, and comes from the proxy addon layer.
func NewConnectionStatsWithDuration(cs ConnectionStatsReaderAdapter, doneAt int64) *ConnectionStats {
	if cs == nil {
		return nil
	}
	logOutput := newConnectionStats(cs)
	logOutput.Duration = doneAt
	return logOutput
}
