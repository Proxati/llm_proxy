package addons

import (
	"log/slog"

	px "github.com/proxati/mitmproxy/proxy"
)

var logger = slog.With("addon", "StdOutLogger")

// StdOutLogger log connection and flow
type StdOutLogger struct {
	px.BaseAddon
}

func (addon *StdOutLogger) ClientConnected(client *px.ClientConn) {
	logger.Info("client connect", "clientAddress", client.Conn.RemoteAddr())
}

func (addon *StdOutLogger) ClientDisconnected(client *px.ClientConn) {
	logger.Info("client disconnect", "clientAddress", client.Conn.RemoteAddr())
}

func (addon *StdOutLogger) ServerConnected(connCtx *px.ConnContext) {
	logger.Info(
		"server connect",
		"serverAddress", connCtx.ServerConn.Address,
		"localAddress", connCtx.ServerConn.Conn.LocalAddr(),
		"remoteAddress", connCtx.ServerConn.Conn.RemoteAddr(),
	)
}

func (addon *StdOutLogger) ServerDisconnected(connCtx *px.ConnContext) {
	logger.Info(
		"server disconnect",
		"serverAddress", connCtx.ServerConn.Address,
		"localAddress", connCtx.ServerConn.Conn.LocalAddr(),
		"remoteAddress", connCtx.ServerConn.Conn.RemoteAddr(),
	)
}

func NewStdOutLogger() *StdOutLogger {
	return &StdOutLogger{}
}
