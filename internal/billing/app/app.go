package app

import (
	"log/slog"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

type HandlerProvider commonhandler.HandlerProvider

func NewHandlerProvider(traceProvider trace.TracerProvider, logger *slog.Logger) *HandlerProvider {
	return (*HandlerProvider)(commonhandler.NewHandlerProvider(
		commonhandler.WithTracer(traceProvider.Tracer("billing-app")),
		commonhandler.WithLogger(logger),
	))
}

var ProvideHandlerProvider = NewHandlerProvider

type Cmds struct{}

func NewCmds(handlerProvider *HandlerProvider) *Cmds {
	return &Cmds{}
}

var ProvideCmds = NewCmds

type Queries struct{}

func NewQueries(handlerProvider *HandlerProvider) *Queries {
	return &Queries{}
}

var ProvideQueries = NewQueries

type Server struct {
	Cmds    *Cmds
	Queries *Queries
}
