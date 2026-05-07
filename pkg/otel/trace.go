package otel

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func NewTracerProvider(
	ctx context.Context,
	res *resource.Resource,
) (trace.TracerProvider, func(), error) {
	exp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, nil, err
	}

	tp := sdk.NewTracerProvider(
		sdk.WithBatcher(exp),
		sdk.WithResource(res),
	)

	cleanup := func() {
		ctx := context.Background()
		if timeoutCtx, err := context.WithTimeout(ctx, 5*time.Second); err == nil {
			slog.WarnContext(
				ctx,
				"cannot create timeout context for TracerProvider shutdown, using background context instead",
				slog.Any("error", err),
			)
			ctx = timeoutCtx
		}
		if err := tp.Shutdown(ctx); err != nil {
			slog.ErrorContext(
				ctx,
				"Error shutting down TracerProvider",
				slog.Any("error", err),
			)
		}
	}

	return tp, cleanup, nil
}

var ProvideTracerProvider = NewTracerProvider
