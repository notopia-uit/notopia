package authorization

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	commongrpc "github.com/notopia-uit/notopia/pkg/common/grpc"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type GRPCServer struct {
	*http.Server
}

func NewGRPCServer(
	ctx context.Context,
	handler pbconnect.AuthorizationServiceHandler,
	cfg *ServerConfig,
	traceProvider *trace.TracerProvider,
	meterProvider *metric.MeterProvider,
	logger *slog.Logger,
) (*GRPCServer, func(), error) {
	interceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(traceProvider),
		otelconnect.WithMeterProvider(meterProvider),
		otelconnect.WithTrustRemote(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create otel interceptor: %w", err)
	}
	errInterceptor := commongrpc.NewErrorInterceptor()
	Path, Handler := pbconnect.NewAuthorizationServiceHandler(
		handler,
		connect.WithInterceptors(interceptor, errInterceptor),
	)
	mux := http.NewServeMux()
	mux.Handle(Path, Handler)
	protocol := new(http.Protocols)
	protocol.SetHTTP1(true)
	protocol.SetUnencryptedHTTP2(true)
	server := &GRPCServer{
		Server: &http.Server{
			Addr:      cfg.GRPC.Address(),
			Handler:   mux,
			Protocols: protocol,
		},
	}
	cleanup := func() {
		if err := server.Shutdown(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown grpc server", slog.String("error", err.Error()))
		}
	}
	return server, cleanup, nil
}

var ProvideGRPCServer = NewGRPCServer

func (s *GRPCServer) Run() error {
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start grpc server: %w", err)
	}
	return nil
}
