package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type ServiceServer struct {
	pb.UnimplementedNoteServiceServer
	app *app.Server
}

var _ pb.NoteServiceServer = (*ServiceServer)(nil)

func NewServiceServer(app *app.Server) *ServiceServer {
	return &ServiceServer{
		app:                            app,
		UnimplementedNoteServiceServer: pb.UnimplementedNoteServiceServer{},
	}
}

var ProvideServiceServer = NewServiceServer

func toGRPCError(err error) error {
	if cerr, ok := errors.AsType[*errs.Err](err); ok {
		switch cerr.Code() {
		case errs.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, cerr.Error())
		case errs.CodeForbidden:
			return status.Error(codes.PermissionDenied, cerr.Error())
		case errs.CodeInvalid,
			errs.CodeEmptyFolderName,
			errs.CodePersistenceInvalid,
			errs.CodeInvalidWorkspaceName,
			errs.CodeInvalidWorkspaceSlug,
			errs.CodeFoldersNotInWorkspace,
			errs.CodeDestinationFolderNotInWorkspace,
			errs.CodeNotesNotInWorkspace,
			errs.CodeWorkspaceMembersCannotBeEmpty,
			errs.CodeWorkspaceMustHaveAtLeastOneOwner:
			return status.Error(codes.InvalidArgument, cerr.Error())
		case errs.CodeUnimplemented:
			return status.Error(codes.Unimplemented, cerr.Error())
		case errs.CodeInternal,
			errs.CodeNoteFailToMarshalDocumentContent,
			errs.CodePersistenceInternal,
			errs.CodeWorkspaceEventPubSubFailedToCreateMessage,
			errs.CodeWorkspaceEventPubSubPublishFailed,
			errs.CodeWorkspaceEventPubSubSubscribeFailed,
			errs.CodeInternalGenerateID:
			return status.Error(codes.Internal, cerr.Error())
		case errs.CodeFolderNotFound,
			errs.CodeNoteNotFound,
			errs.CodeWorkspaceNotFound,
			errs.CodeWorkspaceBySlugNotFound,
			errs.CodeWorkspaceRootFolderNotFound:
			return status.Error(codes.NotFound, cerr.Error())
		case errs.CodeFolderAlreadyExisted,
			errs.CodeFolderAlreadyTrashed,
			errs.CodeNoteAlreadyTrashed,
			errs.CodeWorkspaceSlugAlreadyExists,
			errs.CodeCannotMoveFolderToItOwnSubfolder:
			return status.Error(codes.AlreadyExists, cerr.Error())
		case errs.CodeAuthorizationServiceInternalError:
			return status.Error(codes.Internal, cerr.Error())
		default:
			return status.Error(codes.Unknown, cerr.Error())
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

type GRPC struct {
	server  *grpc.Server
	address string
}

func New(
	ctx context.Context,
	serviceServer pb.NoteServiceServer,
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
