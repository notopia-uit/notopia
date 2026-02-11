package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewMeterProvider(
	ctx context.Context,
	cfg *MeterConfig,
	res *resource.Resource,
) (*sdk.MeterProvider, func(), error) {
	var exporters []sdk.Exporter

	if cfg.Stdout {
		stdoutExp, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, stdoutExp)
	}

	if cfg.GRPC.Endpoint != "" {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(cfg.GRPC.Endpoint),
		}
		if !cfg.GRPC.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		exporter, err := otlpmetricgrpc.New(ctx, opts...)
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, exporter)
	}

	options := []sdk.Option{
		sdk.WithResource(res),
	}

	for _, exporter := range exporters {
		options = append(
			options,
			sdk.WithReader(
				sdk.NewPeriodicReader(
					exporter,
					sdk.WithInterval(cfg.Interval),
				),
			),
		)
	}

	mp := sdk.NewMeterProvider(options...)

	cleanup := func() {
		if err := mp.Shutdown(ctx); err != nil {
			slog.Error(
				"Error shutting down MeterProvider",
				slog.String("error", err.Error()),
			)
		}
	}

	return mp, cleanup, nil
}

var ProvideMeterProvider = NewMeterProvider
