package fileUtils

import "log/slog"

// getLogger avoids a race condition where the logger hasn't yet been configured, so the .Default() is empty
func getLogger() *slog.Logger {
	if _sLogger == nil {
		_sLogger = slog.Default().WithGroup("fileUtils")
	}

	return _sLogger
}

var _sLogger *slog.Logger
