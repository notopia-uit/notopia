package otlp

import (
	"context"
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
) (*sdk.MeterProvider, error) {
	var exporters []sdk.Exporter

	if cfg.MeterStdoutEnabled() {
		stdoutExp, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, err
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
			return nil, err
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

	return sdk.NewMeterProvider(options...), nil
}

var ProvideMeterProvider = NewMeterProvider
