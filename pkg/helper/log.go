package helper

import (
	"errors"
	"log/slog"
)

var ErrInvalidLogLevel = errors.New("invalid log level")

func GetLogLevelFromString(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, ErrInvalidLogLevel
	}
}
