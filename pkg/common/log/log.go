package log

import (
	"log/slog"
	"os"

	"github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
	slogmulti "github.com/samber/slog-multi"
)

type StdoutHandler slog.Handler

func NewStdoutHandler(cfg *config.General) StdoutHandler {
	if cfg.LogDisabled {
		return slog.DiscardHandler
	}
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.GetSlogLevel(),
	})
}

var ProvideStdoutHandler = NewStdoutHandler

func NewMulti(
	stdoutHandler StdoutHandler,
	otelHandler otel.SlogHandler,
) *slog.Logger {
	logger := slog.New(slogmulti.Fanout(stdoutHandler, otelHandler))
	return logger
}

var ProvideMulti = NewMulti
