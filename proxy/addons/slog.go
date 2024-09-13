package addons

import (
	"context"
	"log/slog"

	px "github.com/proxati/mitmproxy/proxy"
)

func getProxyIDSafe(f *px.Flow) string {
	if f.ConnContext != nil && f.ConnContext.ClientConn != nil {
		return f.ConnContext.ID().String()
	}
	return ""
}

func configLoggerFieldsWithFlow(l *slog.Logger, f *px.Flow) *slog.Logger {
	if !l.Enabled(context.TODO(), slog.LevelDebug) || f == nil {
		// Skip if debug is disabled or flow is nil
		return l
	}

	logger := l.With(
		"proxy.ID", f.Id.String(),
		"client.ID", getProxyIDSafe(f),
	)

	if f.Request != nil {
		logger = logger.With(
			slog.Group("Request",
				"URL", f.Request.URL,
				"Method", f.Request.Method,
			),
		)
	}

	if f.Response != nil {
		logger = logger.With(
			slog.Group("Response",
				"StatusCode", f.Response.StatusCode,
			),
		)
	}

	return logger
}
