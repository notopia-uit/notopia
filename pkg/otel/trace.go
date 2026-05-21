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
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(timeoutCtx); err != nil {
			slog.ErrorContext(
				timeoutCtx,
				"Error shutting down TracerProvider",
				slog.Any("error", err),
			)
		}
	}

	return tp, cleanup, nil
}

var ProvideTracerProvider = NewTracerProvider
