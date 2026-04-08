package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func authorizationUnaryClientErrorInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			return errs.NewAuthorizationInternal(err)
		}
		return nil
	}
}

type Authorization struct {
	client pb.AuthorizationServiceClient
}

var _ app.AuthorizationService = (*Authorization)(nil)

func NewAuthorization(
	servicesCfg *config.Services,
	logger logging.Logger,
) (*Authorization, func(), error) {
	statsHandler := otelgrpc.NewClientHandler()
	conn, err := grpc.NewClient(
		servicesCfg.Authorization.URL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(statsHandler),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(logger, logging.WithLogOnEvents(logging.StartCall, logging.FinishCall)),
			authorizationUnaryClientErrorInterceptor(),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial authorization service: %w", err)
	}
	client := pb.NewAuthorizationServiceClient(conn)

	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close authorization service connection", "error", err)
		}
	}
	return &Authorization{
		client: client,
	}, cleanup, nil
}

var ProvideAuthorization = NewAuthorization

func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspacePermission) (bool, error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceItemPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceNotePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceFolderPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) CreateWorkspaceWithOwnership(ctx context.Context, ownerID string, workspaceID uuid.UUID) error {
	return errs.NewUnimplemented()
}

func (a *Authorization) GetWorkspaceMembers(ctx context.Context, userID string, workspaceID uuid.UUID) ([]*app.WorkspaceMemberInfo, error) {
	return nil, errs.NewUnimplemented()
}
