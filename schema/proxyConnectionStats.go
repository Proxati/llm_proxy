package schema

import (
	"encoding/json"

	"github.com/proxati/llm_proxy/v2/schema/proxyadapters"
)

type ProxyConnectionStats struct {
	ClientAddress string `json:"client_address"`
	URL           string `json:"url"`
	Duration      int64  `json:"duration_ms"`
	ProxyID       string `json:"proxy_id,omitempty"`
}

func (obj *ProxyConnectionStats) ToJSON() []byte {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		getLogger().Error("Could not convert object to JSON", "error", err)
		return []byte("{}")
	}
	return jsonData
}

func (obj *ProxyConnectionStats) ToJSONstr() string {
	return string(obj.ToJSON())
}

func newConnectionStats(cs proxyadapters.ConnectionStatsReaderAdapter) *ProxyConnectionStats {
	logOutput := &ProxyConnectionStats{
		// ClientAddress: getClientAddr(f),
		// ProxyID:       f.Id.String(),
		ClientAddress: cs.GetClientIP(),
		ProxyID:       cs.GetProxyID(),
		URL:           cs.GetRequestURL(),
	}

	return logOutput
}

// NewProxyConnectionStatsWithDuration is a slightly leaky abstraction, the doneAt param is for logging
// the entire session length, and comes from the proxy addon layer.
func NewProxyConnectionStatsWithDuration(cs proxyadapters.ConnectionStatsReaderAdapter, doneAt int64) *ProxyConnectionStats {
	if cs == nil {
		return nil
	}
	logOutput := newConnectionStats(cs)
	logOutput.Duration = doneAt
	return logOutput
}
