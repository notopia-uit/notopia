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
