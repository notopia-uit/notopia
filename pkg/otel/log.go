package otel

import (
	"context"
	"log/slog"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/notopia-uit/notopia/pkg/metadata"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/log"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewLoggerProvider(
	ctx context.Context,
	res *resource.Resource,
) (log.LoggerProvider, func(), error) {
	exp, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, nil, err
	}

	lp := sdk.NewLoggerProvider(
		sdk.WithProcessor(sdk.NewBatchProcessor(exp)),
		sdk.WithResource(res),
	)

	cleanup := func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := lp.Shutdown(timeoutCtx); err != nil {
			slog.ErrorContext(
				timeoutCtx,
				"Error shutting down LoggerProvider",
				slog.Any("error", err),
			)
		}
	}

	return lp, cleanup, nil
}

var ProvideLoggerProvider = NewLoggerProvider

type SlogHandler slog.Handler

func NewSlogHandler(
	serviceName metadata.ServiceName,
	provider log.LoggerProvider,
) SlogHandler {
	return otelslog.NewHandler(
		serviceName.String(),
		otelslog.WithLoggerProvider(provider),
	)
}

var ProvideSlogHandler = NewSlogHandler

func MapSlogToGRPCMiddlewareLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		//nolint:sloglint
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
