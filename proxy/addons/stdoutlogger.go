package addons

import (
	"fmt"

	px "github.com/proxati/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// StdOutLogger log connection and flow
type StdOutLogger struct {
	px.BaseAddon
}

func (addon *StdOutLogger) ClientConnected(client *px.ClientConn) {
	go func() {
		log.InfoFn(func() []interface{} {
			return []interface{}{fmt.Sprintf("client connect: %v", client.Conn.RemoteAddr())}
		})
	}()
}

func (addon *StdOutLogger) ClientDisconnected(client *px.ClientConn) {
	go func() {
		log.InfoFn(func() []interface{} {
			return []interface{}{fmt.Sprintf("client disconnect: %v", client.Conn.RemoteAddr())}
		})
	}()
}

func (addon *StdOutLogger) ServerConnected(connCtx *px.ConnContext) {
	go func() {
		log.InfoFn(func() []interface{} {
			return []interface{}{
				fmt.Sprintf("server connect: %v (%v->%v)",
					connCtx.ServerConn.Address,
					connCtx.ServerConn.Conn.LocalAddr(),
					connCtx.ServerConn.Conn.RemoteAddr(),
				),
			}
		})
	}()
}

func (addon *StdOutLogger) ServerDisconnected(connCtx *px.ConnContext) {
	go func() {
		log.InfoFn(func() []interface{} {
			return []interface{}{
				fmt.Sprintf("server disconnect: %v (%v->%v)",
					connCtx.ServerConn.Address,
					connCtx.ServerConn.Conn.LocalAddr(),
					connCtx.ServerConn.Conn.RemoteAddr(),
				),
			}
		})
	}()
}

func NewStdOutLogger() *StdOutLogger {
	return &StdOutLogger{}
}
