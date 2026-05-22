package commonhandler

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Query[Q any, R any] interface {
	Handle(ctx context.Context, query *Q) (R, error)
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
