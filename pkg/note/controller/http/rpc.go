package http

import (
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/notopia-uit/notopia/pkg/note/app"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type IRCPHandler = pbconnect.NoteServiceHandler

type RPCHandler struct {
	app app.App
}

var _ IRCPHandler = (*RPCHandler)(nil)

func NewRPCHandler(app *app.App) *RPCHandler {
	return &RPCHandler{
		app: *app,
	}
}

var ProvideRPCHandler = NewRPCHandler

type (
	RPCHTTPPath    = string
	RPCHTTPHandler = http.Handler
)

type RPCHTTPHandlerRegister struct {
	Path    RPCHTTPPath
	Handler RPCHTTPHandler
}

func NewRPCHTTPHandlerRegister(
	rpcHandler pbconnect.NoteServiceHandler,
	traceProvider *trace.TracerProvider,
	meterProvider *metric.MeterProvider,
) (*RPCHTTPHandlerRegister, error) {
	interceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(traceProvider),
		otelconnect.WithMeterProvider(meterProvider),
		otelconnect.WithTrustRemote(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateOtelInterceptor, err)
	}
	Path, Handler := pbconnect.NewNoteServiceHandler(
		rpcHandler,
		connect.WithInterceptors(interceptor),
	)
	return &RPCHTTPHandlerRegister{
		Path:    Path,
		Handler: Handler,
	}, nil
}

var ProvideRPCHTTPHandlerRegister = NewRPCHTTPHandlerRegister
