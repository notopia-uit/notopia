package otel

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
)

type TraceConfig interface {
	ShouldExportStdout() bool
	ShouldExportGRPC() bool
	GetGRPCRemote() *Remote
	GetSampleRate() float64
}

func NewTracerProvider(
	ctx context.Context,
	cfg TraceConfig,
	res *resource.Resource,
) (*sdk.TracerProvider, func(), error) {
	var exporters []sdk.SpanExporter

	if cfg.ShouldExportStdout() {
		stdoutExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, stdoutExp)
	}

	if cfg.ShouldExportGRPC() {
		remote := cfg.GetGRPCRemote()
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(remote.Endpoint),
		}
		if !remote.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, exporter)
	}

	sampler := sdk.ParentBased(sdk.TraceIDRatioBased(cfg.GetSampleRate()))

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
