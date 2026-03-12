package logging

import (
	"log/slog"
	"os"
)

type StdoutHandler slog.Handler

func NewStdoutHandler(cfg *Config) StdoutHandler {
	if !cfg.Enabled {
		return slog.DiscardHandler
	}
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.GetSlogLevel(),
	})
}

var ProvideStdoutHandler = NewStdoutHandler
