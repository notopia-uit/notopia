package logging

import (
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/otel"
	slogmulti "github.com/samber/slog-multi"
)

func New(
	stdoutHandler StdoutHandler,
	otelHandler otel.SlogHandler,
) *slog.Logger {
	logger := slog.New(slogmulti.Fanout(stdoutHandler, otelHandler))
	return logger
}

var Provide = New
