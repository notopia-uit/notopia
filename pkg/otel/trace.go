package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(
	ctx context.Context,
	res *resource.Resource,
) (*sdk.TracerProvider, func(), error) {
	exp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, nil, err
	}

	tp := sdk.NewTracerProvider(
		sdk.WithBatcher(exp),
		sdk.WithResource(res),
	)

	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error(
				"Error shutting down TracerProvider",
				slog.Any("error", err),
			)
		}
	}

	return tp, cleanup, nil
}

var ProvideTracerProvider = NewTracerProvider
