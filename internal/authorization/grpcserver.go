package authorization

import (
	"context"
	"errors"
	"fmt"
	"net"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	if err, ok := errors.AsType[*errs.Err](err); ok {
		switch err.Code() {
		case errs.CodeCasbinInternalError,
			errs.CodeCasbinEnforcerError,
			errs.CodeGetWorkspaceMembersGetFailed,
			errs.CodeInternal:
			return status.Error(codes.Internal, err.Message())
		case errs.CodeCasbinPolicySignatureInvalid,
			errs.CodeErrInvalidUserFormat,
			errs.CodeInvalid:
			return status.Error(codes.InvalidArgument, err.Message())
		case errs.CodeMemberHasNoPermission,
			errs.CodeForbidden:
			return status.Error(codes.PermissionDenied, err.Message())
		case errs.CodeCreateWorkspaceExists:
			return status.Error(codes.AlreadyExists, err.Message())
		case errs.CodeUnimplemented:
			return status.Error(codes.Unimplemented, err.Message())
		}
	}
	return err
}

func unaryErrorInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return resp, nil
}

type GRPCServer struct {
	server  *grpc.Server
	address string
}

func NewGRPCServer(
	ctx context.Context,
	serviceServer pb.AuthorizationServiceServer,
	cfg *ServerConfig,
	logger logging.Logger,
) (*GRPCServer, func(), error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create protovalidate validator: %w", err)
	}
	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(logger, logging.WithLogOnEvents(logging.StartCall, logging.FinishCall)),
			protovalidate_middleware.UnaryServerInterceptor(validator),
			unaryErrorInterceptor,
		),
	)
	grpcServer := &GRPCServer{
		server:  server,
		address: cfg.GRPC.Address(),
	}
	pb.RegisterAuthorizationServiceServer(server, serviceServer)
	cleanup := func() {
		server.GracefulStop()
	}
	return grpcServer, cleanup, nil
}

var ProvideGRPCServer = NewGRPCServer

func (g *GRPCServer) Run() error {
	lis, err := net.Listen("tcp", g.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", g.address, err)
	}
	return g.server.Serve(lis)
}

func (g *GRPCServer) Stop() {
	g.server.GracefulStop()
}
