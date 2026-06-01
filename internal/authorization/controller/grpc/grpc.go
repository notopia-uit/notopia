package grpc

import (
	"context"
	"fmt"
	"net"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/notopia-uit/notopia/internal/authorization/config"
	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type Server struct {
	server  *grpc.Server
	address string
}

func NewServer(
	ctx context.Context,
	serviceServer *Service,
	cfg *config.ServerConfig,
	logger logging.Logger,
	_ otel.Global,
) (*Server, func(), error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create protovalidate validator: %w", err)
	}
	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
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
	grpcServer := &Server{
		server:  server,
		address: cfg.GRPC.Address(),
	}
	pb.RegisterAuthorizationServiceServer(
		server,
		serviceServer,
	)
	return grpcServer, server.GracefulStop, nil
}

var ProvideServer = NewServer

func (g *Server) Run() error {
	lis, err := net.Listen("tcp", g.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", g.address, err)
	}
	return g.server.Serve(lis)
}

func (g *Server) GracefulStop() {
	g.server.GracefulStop()
}

func (g *Server) Stop() {
	g.server.Stop()
}
