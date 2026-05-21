package commonhandler

import (
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

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
