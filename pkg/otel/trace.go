package otel

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(
	ctx context.Context,
	cfg *TraceConfig,
	res *resource.Resource,
) (*sdk.TracerProvider, func(), error) {
	var exporters []sdk.SpanExporter

	if cfg.Stdout {
		stdoutExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, stdoutExp)
	}

	if cfg.GRPC.Endpoint != "" {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.GRPC.Endpoint),
		}
		if cfg.GRPC.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, exporter)
	}

	sampler := sdk.ParentBased(sdk.TraceIDRatioBased(cfg.SampleRate))

	options := []sdk.TracerProviderOption{
		sdk.WithResource(res),
		sdk.WithSampler(sampler),
	}

	for _, exporter := range exporters {
		options = append(options, sdk.WithBatcher(exporter))
	}

	tp := sdk.NewTracerProvider(options...)

	cleanup := func() {
		_ = tp.Shutdown(ctx)
		for _, exporter := range exporters {
			_ = exporter.Shutdown(ctx)
		}
	}

	return tp, cleanup, nil
}

var ProvideTracerProvider = NewTracerProvider
