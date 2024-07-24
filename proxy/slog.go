package proxy

import "log/slog"

var sLogger *slog.Logger = slog.Default().WithGroup("proxy")
