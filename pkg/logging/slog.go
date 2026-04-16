package logging

import (
	"context"
	"log/slog"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
	slogmulti "github.com/samber/slog-multi"
)

func NewSlog(
	stdoutHandler StdoutHandler,
	otelHandler otel.SlogHandler,
	cfg *commonconfig.Log,
) *slog.Logger {
	minLevel := cfg.GetSlogLevel()
	middleware := slogmulti.NewEnabledInlineMiddleware(func(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool {
		if level < minLevel {
			return false
		}
		return next(ctx, level)
	})
	return slog.New(
		slogmulti.
			Pipe(middleware).
			Handler(slogmulti.Fanout(stdoutHandler, otelHandler)),
	)
}

var ProvideSlog = NewSlog
