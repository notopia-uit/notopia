package commonhandler

import (
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type Decorator struct {
	tracer trace.Tracer
	logger *slog.Logger
}

type DecoratorOption func(*Decorator)

func WithTracer(tracer trace.Tracer) DecoratorOption {
	return func(d *Decorator) {
		d.tracer = tracer
	}
}

func WithLogger(logger *slog.Logger) DecoratorOption {
	return func(d *Decorator) {
		d.logger = logger
	}
}

func NewDecorator(options ...DecoratorOption) *Decorator {
	d := &Decorator{}
	for _, option := range options {
		option(d)
	}
	return d
}

type decorateConfig struct {
	skipTrace bool
	skipLog   bool
}

type DecorateOption func(*decorateConfig)

func WithoutTrace() DecorateOption {
	return func(c *decorateConfig) {
		c.skipTrace = true
	}
}

func WithoutLog() DecorateOption {
	return func(c *decorateConfig) {
		c.skipLog = true
	}
}

// Wrappers are applied inside-out: log runs within the trace span so log records carry the trace ID.

func (d *Decorator) Cmd[C any](handler Cmd[C], options ...DecorateOption) Cmd[C] {
	var c decorateConfig
	for _, option := range options {
		option(&c)
	}
	handlerName := fmt.Sprintf("%T", handler)
	decorated := handler
	if d.logger != nil && !c.skipLog {
		decorated = NewLogCmd(decorated, d.logger, handlerName)
	}
	if d.tracer != nil && !c.skipTrace {
		decorated = NewTraceCmd(decorated, d.tracer, handlerName)
	}
	return decorated
}

func (d *Decorator) Query[Q, R any](handler Query[Q, R], options ...DecorateOption) Query[Q, R] {
	var c decorateConfig
	for _, option := range options {
		option(&c)
	}
	handlerName := fmt.Sprintf("%T", handler)
	decorated := handler
	if d.logger != nil && !c.skipLog {
		decorated = NewLogQuery(decorated, d.logger, handlerName)
	}
	if d.tracer != nil && !c.skipTrace {
		decorated = NewTraceQuery(decorated, d.tracer, handlerName)
	}
	return decorated
}
