package otlp

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/common/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(ctx context.Context, cfg *config.OTLP, res *resource.Resource) (*sdk.TracerProvider, error) {
	var exporters []sdk.SpanExporter

	if cfg.TraceStdoutEnabled() {
		stdoutExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, stdoutExp)
	}

	if remoteCfg, ok := cfg.TraceEndpoint(); ok {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(remoteCfg.Endpoint),
		}
		if !remoteCfg.InSecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, exporter)
	}

	sampler := sdk.ParentBased(sdk.TraceIDRatioBased(cfg.Trace.SampleRate))

	options := []sdk.TracerProviderOption{
		sdk.WithResource(res),
		sdk.WithSampler(sampler),
	}

	for _, exporter := range exporters {
		options = append(options, sdk.WithBatcher(exporter))
	}

	return sdk.NewTracerProvider(options...), nil
}
