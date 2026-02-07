package otlp

import (
	"context"
	"log"
	"time"

	"github.com/notopia-uit/notopia/pkg/common/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewMeterProvider(
	ctx context.Context,
	cfg *config.OTLP,
	res *resource.Resource,
) (*sdk.MeterProvider, func(), error) {
	var exporters []sdk.Exporter

	if cfg.MeterStdoutEnabled() {
		stdoutExp, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, stdoutExp)
	}

	if remoteCfg, ok := cfg.MeterEndpoint(); ok {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(remoteCfg.Endpoint),
		}
		if !remoteCfg.InSecure {
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
					sdk.WithInterval(
						cfg.MetricExportInterval()*time.Second,
					),
				),
			),
		)
	}

	mp := sdk.NewMeterProvider(options...)

	cleanup := func() {
		if err := mp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down MeterProvider: %v", err)
		}
	}

	return mp, cleanup, nil
}

var ProvideMeterProvider = NewMeterProvider
