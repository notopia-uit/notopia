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

type ServiceServerAdapter struct {
	pb.UnimplementedNoteServiceServer
	NoteServiceServer *ServiceServer
}

func NewServiceServerAdapter(serviceServer *ServiceServer) *ServiceServerAdapter {
	return &ServiceServerAdapter{
		UnimplementedNoteServiceServer: pb.UnimplementedNoteServiceServer{},
		NoteServiceServer:              serviceServer,
	}
}

var ProvideServiceServerAdapter = NewServiceServerAdapter

type GRPC struct {
	server  *grpc.Server
	address string
}

func New(
	ctx context.Context,
	serviceServer *ServiceServerAdapter,
	cfg *config.Server,
	logger logging.Logger,
) (*GRPC, func(), error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create protovalidate validator: %w", err)
	}
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(logger, logging.WithLogOnEvents(logging.StartCall, logging.FinishCall)),
			protovalidate_middleware.UnaryServerInterceptor(validator),
			unaryErrorInterceptor,
		),
	)
	grpc := &GRPC{
		server:  grpcServer,
		address: cfg.GRPC.Address(),
	}
	pb.RegisterNoteServiceServer(grpcServer, serviceServer)
	cleanup := func() {
		grpcServer.GracefulStop()
	}
	return grpc, cleanup, nil
}

func (g *GRPC) Run() error {
	lis, err := net.Listen("tcp", g.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", g.address, err)
	}
	return g.server.Serve(lis)
}

func (g *GRPC) Stop() {
	g.server.GracefulStop()
}

var Provide = New
