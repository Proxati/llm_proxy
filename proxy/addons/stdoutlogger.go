package addons

import (
	"log/slog"

	px "github.com/proxati/mitmproxy/proxy"
)

// StdOutLogger log connection and flow
type StdOutLogger struct {
	px.BaseAddon
	logger *slog.Logger
}

func (a *StdOutLogger) ClientConnected(client *px.ClientConn) {
	a.logger.Debug(
		"client connect",
		"clientAddress", client.Conn.RemoteAddr(),
		"ID", client.ID,
	)
}

func (a *StdOutLogger) ClientDisconnected(client *px.ClientConn) {
	a.logger.Debug(
		"client disconnect",
		"clientAddress", client.Conn.RemoteAddr(),
		"ID", client.ID,
	)
}

func (a *StdOutLogger) ServerConnected(connCtx *px.ConnContext) {
	a.logger.Debug(
		"server connect",
		"serverAddress", connCtx.ServerConn.Address,
		"localAddress", connCtx.ServerConn.Conn.LocalAddr(),
		"remoteAddress", connCtx.ServerConn.Conn.RemoteAddr(),
		"ID", connCtx.ID(),
	)
}

func (a *StdOutLogger) ServerDisconnected(connCtx *px.ConnContext) {
	a.logger.Debug(
		"server disconnect",
		"serverAddress", connCtx.ServerConn.Address,
		"localAddress", connCtx.ServerConn.Conn.LocalAddr(),
		"remoteAddress", connCtx.ServerConn.Conn.RemoteAddr(),
		"ID", connCtx.ID(),
	)
}

func NewStdOutLogger(logger *slog.Logger) *StdOutLogger {
	return &StdOutLogger{
		logger: logger.WithGroup("addons.StdOutLogger"),
	}
}
