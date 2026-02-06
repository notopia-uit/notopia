package otlp

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/common/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func InitLogExporter(ctx context.Context, cfg *config.OTLP, res *resource.Resource) (*sdk.LoggerProvider, error) {
	var exporters []sdk.Exporter

	if cfg.LogStdoutEnabled() {
		stdoutExp, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, stdoutExp)
	}

	if remoteCfg, ok := cfg.LogEndpoint(); ok {
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(remoteCfg.Endpoint),
		}
		if !remoteCfg.InSecure {
			opts = append(opts, otlploggrpc.WithInsecure())
		}
		exporter, err := otlploggrpc.New(ctx, opts...)
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, exporter)
	}

	otions := []sdk.LoggerProviderOption{
		sdk.WithResource(res),
	}

	for _, exporter := range exporters {
		otions = append(
			otions,
			sdk.WithProcessor(sdk.NewBatchProcessor(exporter)),
		)
	}

	return sdk.NewLoggerProvider(otions...), nil
}
