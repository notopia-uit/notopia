package otel

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/pkg/common/metadata"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// HACK: This struct created by AI
type severityFilter struct {
	sdk.Processor
	minSeverity log.Severity
}

func (f *severityFilter) Enabled(_ context.Context, param sdk.EnabledParameters) bool {
	return param.Severity >= f.minSeverity
}

func (f *severityFilter) OnEmit(ctx context.Context, record *sdk.Record) error {
	if record.Severity() < f.minSeverity {
		return nil
	}
	return f.Processor.OnEmit(ctx, record)
}

type LogConfig interface {
	ShouldExportStdout() bool
	ShouldExportGRPC() bool
	GetGRPCRemote() *Remote
	GetMinSecurity() log.Severity
}

func NewLoggerProvider(
	ctx context.Context,
	cfg LogConfig,
	res *resource.Resource,
) (*sdk.LoggerProvider, func(), error) {
	var exporters []sdk.Exporter

	if cfg.ShouldExportStdout() {
		stdoutExp, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, stdoutExp)
	}

	if cfg.ShouldExportGRPC() {
		remote := cfg.GetGRPCRemote()
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(remote.Endpoint),
		}
		if !remote.Insecure {
			opts = append(opts, otlploggrpc.WithInsecure())
		}
		exporter, err := otlploggrpc.New(ctx, opts...)
		if err != nil {
			return nil, nil, err
		}
		exporters = append(exporters, exporter)
	}

	options := []sdk.LoggerProviderOption{
		sdk.WithResource(res),
	}

	for _, exporter := range exporters {
		processor := sdk.NewBatchProcessor(exporter)
		filter := &severityFilter{
			minSeverity: cfg.GetMinSecurity(),
			Processor:   processor,
		}

		options = append(
			options,
			sdk.WithProcessor(filter),
		)
	}

	lp := sdk.NewLoggerProvider(options...)

	cleanup := func() {
		_ = lp.Shutdown(ctx)
	}

	return lp, cleanup, nil
}

var ProvideLoggerProvider = NewLoggerProvider

func NewSlog(
	serviceName metadata.ServiceName,
	provider *sdk.LoggerProvider,
) *slog.Logger {
	logger := otelslog.NewLogger(
		serviceName.String(),
		otelslog.WithLoggerProvider(provider),
	)
	return logger
}

var ProvideSlog = NewSlog
