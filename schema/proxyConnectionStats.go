package schema

import (
	"encoding/json"
)

const UnknownAddr = "unknown"

type ProxyConnectionStats struct {
	ClientAddress string `json:"client_address"`
	URL           string `json:"url"`
	Duration      int64  `json:"duration_ms"`
	ProxyID       string `json:"proxy_id,omitempty"`
}

func (obj *ProxyConnectionStats) toJSON() []byte {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		getLogger().Error("Could not convert object to JSON", "error", err)
		return []byte("{}")
	}
	return jsonData
}

func (obj *ProxyConnectionStats) toJSONstr() string {
	return string(obj.toJSON())
}

func newConnectionStats(cs ConnectionStatsReaderAdapter) *ProxyConnectionStats {
	logOutput := &ProxyConnectionStats{
		// ClientAddress: getClientAddr(f),
		// ProxyID:       f.Id.String(),
		ClientAddress: cs.GetClientIP(),
		ProxyID:       cs.GetProxyID(),
		URL:           cs.GetRequestURL(),
	}

	return logOutput
}

// newProxyConnectionStatsWithDuration is a slightly leaky abstraction, the doneAt param is for logging
// the entire session length, and comes from the proxy addon layer.
func newProxyConnectionStatsWithDuration(cs ConnectionStatsReaderAdapter, doneAt int64) *ProxyConnectionStats {
	if cs == nil {
		return nil
	}
	logOutput := newConnectionStats(cs)
	logOutput.Duration = doneAt
	return logOutput
}
