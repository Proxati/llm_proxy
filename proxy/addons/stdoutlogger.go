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
	a.logger.Info("client connect", "clientAddress", client.Conn.RemoteAddr())
}

func (a *StdOutLogger) ClientDisconnected(client *px.ClientConn) {
	a.logger.Info("client disconnect", "clientAddress", client.Conn.RemoteAddr())
}

func (a *StdOutLogger) ServerConnected(connCtx *px.ConnContext) {
	a.logger.Info(
		"server connect",
		"serverAddress", connCtx.ServerConn.Address,
		"localAddress", connCtx.ServerConn.Conn.LocalAddr(),
		"remoteAddress", connCtx.ServerConn.Conn.RemoteAddr(),
	)
}

func (a *StdOutLogger) ServerDisconnected(connCtx *px.ConnContext) {
	a.logger.Info(
		"server disconnect",
		"serverAddress", connCtx.ServerConn.Address,
		"localAddress", connCtx.ServerConn.Conn.LocalAddr(),
		"remoteAddress", connCtx.ServerConn.Conn.RemoteAddr(),
	)
}

func NewStdOutLogger(logger *slog.Logger) *StdOutLogger {
	return &StdOutLogger{
		logger: logger.WithGroup("addons").With("name", "StdOutLogger"),
	}
}
