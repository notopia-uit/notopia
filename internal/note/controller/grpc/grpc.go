package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	commongrpc "github.com/notopia-uit/notopia/pkg/common/grpc"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type IHandler = pbconnect.NoteServiceHandler

type Handler struct {
	app app.App
}

var _ IHandler = (*Handler)(nil)

func NewHandler(app *app.App) *Handler {
	return &Handler{
		app: *app,
	}
}

var ProvideHandler = NewHandler

type Server struct {
	*http.Server
}

func New(
	ctx context.Context,
	handler IHandler,
	cfg *config.Server,
	traceProvider *trace.TracerProvider,
	meterProvider *metric.MeterProvider,
	logger *slog.Logger,
) (*Server, func(), error) {
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(traceProvider),
		otelconnect.WithMeterProvider(meterProvider),
		otelconnect.WithTrustRemote(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create otel interceptor: %w", err)
	}
	validateInterceptor := validate.NewInterceptor()
	// TODO: interceptor from authorization service
	// When I refactor the authorization service later
	errInterceptor := commongrpc.NewErrorInterceptor()
	Path, Handler := pbconnect.NewNoteServiceHandler(
		handler,
		connect.WithInterceptors(
			otelInterceptor,
			validateInterceptor,
			errInterceptor,
		),
	)
	mux := http.NewServeMux()
	mux.Handle(Path, Handler)
	protocol := new(http.Protocols)
	protocol.SetHTTP1(true)
	protocol.SetUnencryptedHTTP2(true)
	server := &Server{
		Server: &http.Server{
			Addr:      cfg.GRPC.Address(),
			Handler:   mux,
			Protocols: protocol,
		},
	}
	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown grpc server", slog.String("error", err.Error()))
		}
	}
	return server, cleanup, nil
}

func (s *Server) Run() error {
	return s.ListenAndServe()
}

var Provide = New
