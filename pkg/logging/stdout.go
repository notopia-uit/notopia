package logging

import (
	"log/slog"
	"os"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type StdoutHandler slog.Handler

func NewStdoutHandler(cfg *commonconfig.General) StdoutHandler {
	if cfg.LogDisabled {
		return slog.DiscardHandler
	}
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.GetSlogLevel(),
	})
}

var ProvideStdoutHandler = NewStdoutHandler
