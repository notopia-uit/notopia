package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewMeterProvider(
	ctx context.Context,
	res *resource.Resource,
) (*sdk.MeterProvider, func(), error) {
	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, nil, err
	}
	mp := sdk.NewMeterProvider(
		sdk.WithResource(res),
		sdk.WithReader(reader),
	)

	cleanup := func() {
		if err := mp.Shutdown(ctx); err != nil {
			slog.Error(
				"Error shutting down MeterProvider",
				slog.Any("error", err),
			)
		}
	}

	return mp, cleanup, nil
}

var ProvideMeterProvider = NewMeterProvider
