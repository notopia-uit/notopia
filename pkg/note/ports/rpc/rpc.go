package rpc

import (
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/notopia-uit/notopia/pkg/note/app"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type RPCHandler struct {
	app app.App
}

var _ pbconnect.NoteServiceHandler = (*RPCHandler)(nil)

func NewRPCHandler(app *app.App) *RPCHandler {
	return &RPCHandler{
		app: *app,
	}
}

var ProvideRPCHandler = NewRPCHandler

type (
	HTTPPath    = string
	HTTPHandler = http.Handler
)

type HTTPServiceHandler struct {
	Path    HTTPPath
	Handler HTTPHandler
}

var ErrCreateOtelInterceptor = errors.New("failed to create otel connect interceptor")

func NewHTTPServiceHandler(
	rpcHandler pbconnect.NoteServiceHandler,
	traceProvider *trace.TracerProvider,
	logProvider *log.LoggerProvider,
	meterProvider *metric.MeterProvider,
) (*HTTPServiceHandler, error) {
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
	return &HTTPServiceHandler{
		Path:    Path,
		Handler: Handler,
	}, nil
}

var ProvideHTTPServiceHandler = NewHTTPServiceHandler
