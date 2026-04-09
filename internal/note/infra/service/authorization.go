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

func (a *Authorization) CreateWorkspaceWithOwner(ctx context.Context, ownerID string, workspaceID uuid.UUID) error {
	_, err := a.client.CreateWorkspaceWithOwner(ctx, &pb.CreateWorkspaceWithOwnerRequest{
		OwnerId:     ownerID,
		WorkspaceId: workspaceID.String(),
	})
	return err
}

func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspacePermission) (bool, error) {
	perm, err := a.toWorkspacePermissionProto(permission)
	if err != nil {
		return false, err
	}
	response, err := a.client.HasWorkspacePermission(ctx, &pb.HasWorkspacePermissionRequest{
		MemberId:    userID,
		WorkspaceId: workspaceID.String(),
		Permission:  perm,
	})
	if err != nil {
		return false, err
	}
	return response.HasPermission, nil
}

func (a *Authorization) HasWorkspaceItemPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, error) {
	perm, err := a.toWorkspaceItemPermissionProto(permission)
	if err != nil {
		return false, err
	}
	response, err := a.client.HasWorkspaceItemPermission(ctx, &pb.HasWorkspaceItemPermissionRequest{
		MemberId:    userID,
		WorkspaceId: workspaceID.String(),
		Permission:  perm,
	})
	if err != nil {
		return false, err
	}
	return response.HasPermission, nil
}

func (a *Authorization) UpdateWorkspaceMembers(
	ctx context.Context,
	userID string,
	workspaceID uuid.UUID,
	members []app.WorkspaceMemberUpdate,
) error {
	return errs.NewUnimplemented()
}

func (a *Authorization) GetWorkspaceMembers(ctx context.Context, userID string, workspaceID uuid.UUID) ([]*app.WorkspaceMemberInfo, error) {
	return nil, errs.NewUnimplemented()
}

func (a *Authorization) toWorkspacePermissionProto(permission app.WorkspacePermission) (pb.WorkspacePermission, error) {
	switch permission {
	case app.WorkspacePermissionRead:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_READ, nil
	case app.WorkspacePermissionEdit:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_EDIT, nil
	case app.WorkspacePermissionDelete:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_DELETE, nil
	case app.WorkspacePermissionUnspecified:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_UNSPECIFIED, nil
	default:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace permission: %v", permission), nil)
	}
}

func (a *Authorization) toWorkspaceItemPermissionProto(permission app.WorkspaceItemPermission) (pb.WorkspaceItemPermission, error) {
	switch permission {
	case app.WorkspaceItemPermissionRead:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_READ, nil
	case app.WorkspaceItemPermissionWrite:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_WRITE, nil
	case app.WorkspaceItemPermissionDelete:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_DELETE, nil
	case app.WorkspaceItemPermissionUnspecified:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_UNSPECIFIED, nil
	default:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace item permission: %v", permission), nil)
	}
}
