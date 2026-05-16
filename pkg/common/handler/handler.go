package commonhandler

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Cmd[C any] interface {
	Handle(ctx context.Context, cmd *C) error
}

type Query[Q any, R any] interface {
	Handle(ctx context.Context, query *Q) (R, error)
}

type LogCmd[C any] struct {
	Cmd[C]
	logger      *slog.Logger
	handlerName string
}

var _ Cmd[any] = (*LogCmd[any])(nil)

func NewLogCmd[C any](
	cmd Cmd[C],
	logger *slog.Logger,
	handlerName string,
) *LogCmd[C] {
	return &LogCmd[C]{
		Cmd:         cmd,
		logger:      logger,
		handlerName: handlerName,
	}
}

func (lc *LogCmd[C]) Handle(ctx context.Context, cmd *C) error {
	lc.logger.InfoContext(
		ctx, "Handling command",
		slog.String("handler", lc.handlerName),
		slog.Any("cmd", cmd),
	)
	err := lc.Cmd.Handle(ctx, cmd)
	if err != nil {
		lc.logger.WarnContext(
			ctx, "Error handling command",
			slog.String("handler", lc.handlerName),
			slog.Any("cmd", cmd),
			slog.Any("error", err),
		)
	} else {
		lc.logger.InfoContext(
			ctx, "Successfully handled command",
			slog.String("handler", lc.handlerName),
			slog.Any("cmd", cmd),
		)
	}
	return err
}

type LogQuery[Q any, R any] struct {
	Query[Q, R]
	logger      *slog.Logger
	handlerName string
}

var _ Query[any, any] = (*LogQuery[any, any])(nil)

func NewLogQuery[Q any, R any](
	query Query[Q, R],
	logger *slog.Logger,
	handlerName string,
) *LogQuery[Q, R] {
	return &LogQuery[Q, R]{
		Query:       query,
		logger:      logger,
		handlerName: handlerName,
	}
}

func (lq *LogQuery[Q, R]) Handle(ctx context.Context, query *Q) (R, error) {
	lq.logger.InfoContext(
		ctx, "Handling query",
		slog.String("handler", lq.handlerName),
		slog.Any("query", query),
	)
	result, err := lq.Query.Handle(ctx, query)
	if err != nil {
		lq.logger.WarnContext(
			ctx, "Error handling query",
			slog.String("handler", lq.handlerName),
			slog.Any("query", query),
			slog.Any("error", err),
		)
	} else {
		lq.logger.InfoContext(
			ctx, "Successfully handled query",
			slog.String("handler", lq.handlerName),
			slog.Any("query", query),
		)
		lq.logger.DebugContext(
			ctx, "Query result",
			slog.String("handler", lq.handlerName),
			slog.Any("query", query),
			slog.Any("result", result),
		)
	}
	return result, err
}

type TraceCmd[C any] struct {
	Cmd[C]
	tracer      trace.Tracer
	handlerName string
}

var _ Cmd[any] = (*TraceCmd[any])(nil)

func NewTraceCmd[C any](
	cmd Cmd[C],
	tracer trace.Tracer,
	handlerName string,
) *TraceCmd[C] {
	return &TraceCmd[C]{
		Cmd:         cmd,
		tracer:      tracer,
		handlerName: handlerName,
	}
}

func (tc *TraceCmd[C]) Handle(ctx context.Context, cmd *C) error {
	ctx, span := tc.tracer.Start(ctx, fmt.Sprintf("Cmd.Handle.%s", tc.handlerName))
	defer span.End()
	err := tc.Cmd.Handle(ctx, cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

type TraceQuery[Q any, R any] struct {
	Query[Q, R]
	tracer      trace.Tracer
	handlerName string
}

var _ Query[any, any] = (*TraceQuery[any, any])(nil)

func NewTraceQuery[Q any, R any](
	query Query[Q, R],
	tracer trace.Tracer,
	handlerName string,
) *TraceQuery[Q, R] {
	return &TraceQuery[Q, R]{
		Query:       query,
		tracer:      tracer,
		handlerName: handlerName,
	}
}

func (tq *TraceQuery[Q, R]) Handle(ctx context.Context, query *Q) (R, error) {
	ctx, span := tq.tracer.Start(ctx, fmt.Sprintf("Query.Handle.%s", tq.handlerName))
	defer span.End()
	result, err := tq.Query.Handle(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

type HandlerProvider struct {
	tracer trace.Tracer
	logger *slog.Logger
}

type HandlerProviderOption func(*HandlerProvider)

func WithTracer(tracer trace.Tracer) HandlerProviderOption {
	return func(hp *HandlerProvider) {
		hp.tracer = tracer
	}
}

func WithLogger(logger *slog.Logger) HandlerProviderOption {
	return func(hp *HandlerProvider) {
		hp.logger = logger
	}
}

func NewHandlerProvider(options ...HandlerProviderOption) *HandlerProvider {
	hp := &HandlerProvider{}
	for _, option := range options {
		option(hp)
	}
	return hp
}

// Until https://github.com/golang/go/issues/77273 is resolved
// We will use generic method on HandlerProvider to decorate handlers with tracing and logging

func DecorateCmd[C any](hp *HandlerProvider, handler Cmd[C]) Cmd[C] {
	handlerName := fmt.Sprintf("%T", handler)
	decorated := handler
	if hp.tracer != nil {
		decorated = NewTraceCmd(decorated, hp.tracer, handlerName)
	}
	if hp.logger != nil {
		decorated = NewLogCmd(decorated, hp.logger, handlerName)
	}
	return decorated
}

func DecorateQuery[Q any, R any](hp *HandlerProvider, handler Query[Q, R]) Query[Q, R] {
	handlerName := fmt.Sprintf("%T", handler)
	decorated := handler
	if hp.tracer != nil {
		decorated = NewTraceQuery(decorated, hp.tracer, handlerName)
	}
	if hp.logger != nil {
		decorated = NewLogQuery(decorated, hp.logger, handlerName)
	}
	return decorated
}
