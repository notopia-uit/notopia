package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type ServiceServer struct {
	app *app.Server
}

var _ pb.NoteServiceServer = (*ServiceServer)(nil)

func NewServiceServer(app *app.Server) *ServiceServer {
	return &ServiceServer{
		app: app,
	}
}

var ProvideServiceServer = NewServiceServer

type GRPC struct {
	server  *grpc.Server
	address string
}

func New(
	ctx context.Context,
	serviceServer *ServiceServer,
	cfg *config.Server,
	logger logging.Logger,
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
	propagator propagation.TextMapPropagator,
) (*GRPC, func(), error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create protovalidate validator: %w", err)
	}
	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(tracerProvider),
			otelgrpc.WithMeterProvider(meterProvider),
			otelgrpc.WithPropagators(propagator),
		)),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(logger, logging.WithLogOnEvents(
				logging.StartCall,
				logging.FinishCall,
				logging.PayloadSent,
				logging.PayloadReceived,
			)),
			protovalidate_middleware.UnaryServerInterceptor(validator),
			unaryErrorInterceptor,
		),
	)
	grpc := &GRPC{
		server:  server,
		address: cfg.GRPC.Address(),
	}
	pb.RegisterNoteServiceServer(server, serviceServer)
	return grpc, server.GracefulStop, nil
}

func (g *GRPC) Run() error {
	lis, err := net.Listen("tcp", g.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", g.address, err)
	}
	return g.server.Serve(lis)
}

func (g *GRPC) GracefulStop() {
	g.server.GracefulStop()
}

func (g *GRPC) Stop() {
	g.server.Stop()
}

var Provide = New
