package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
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
			if method == "/grpc.health.v1.Health/Check" {
				return err
			}
			return errs.NewAuthorizationInternal(err)
		}
		return nil
	}
}

type Authorization struct {
	client pb.AuthorizationServiceClient
	conn   *grpc.ClientConn
}

var _ app.AuthorizationSvc = (*Authorization)(nil)

func NewAuthorization(
	servicesCfg *config.Services,
	logger logging.Logger,
) (*Authorization, func(), error) {
	statsHandler := otelgrpc.NewClientHandler()
	logInterceptor := selector.UnaryClientInterceptor(
		logging.UnaryClientInterceptor(logger, logging.WithLogOnEvents(logging.StartCall, logging.FinishCall)),
		selector.MatchFunc(func(ctx context.Context, callMeta interceptors.CallMeta) bool {
			return callMeta.Service != "grpc.health.v1.Health"
		}),
	)
	conn, err := grpc.NewClient(
		servicesCfg.Authorization.URL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(statsHandler),
		grpc.WithChainUnaryInterceptor(logInterceptor, authorizationUnaryClientErrorInterceptor()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid authorization service client configuration: %w", err)
	}
	client := pb.NewAuthorizationServiceClient(conn)

	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close authorization service connection", "error", err)
		}
	}
	return &Authorization{
		client: client,
		conn:   conn,
	}, cleanup, nil
}

var ProvideAuthorization = NewAuthorization

func (a *Authorization) GetUserWorkspaces(ctx context.Context, userID string) ([]app.AuthorizationUserWorkspace, error) {
	response, err := a.client.GetUserWorkspaces(ctx, &pb.GetUserWorkspacesRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	workspaces := make([]app.AuthorizationUserWorkspace, 0, len(response.Workspaces))
	for _, w := range response.Workspaces {
		id, err := uuid.Parse(w.Id)
		if err != nil {
			return nil, errs.NewAuthorizationInternal(fmt.Errorf("invalid workspace ID: %w", err))
		}
		role, err := a.toAppWorkspaceRole(w.Role)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, app.AuthorizationUserWorkspace{
			ID:   id,
			Role: role,
		})
	}
	return workspaces, nil
}

func (a *Authorization) CreateWorkspaceWithOwner(ctx context.Context, ownerID string, workspaceID uuid.UUID) error {
	_, err := a.client.CreateWorkspaceWithOwner(ctx, &pb.CreateWorkspaceWithOwnerRequest{
		OwnerId:     ownerID,
		WorkspaceId: workspaceID.String(),
	})
	return err
}

func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspacePermission) (bool, error) {
	perm, err := a.toAppWorkspacePermission(permission)
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
	perm, err := a.toAppWorkspaceItemPermission(permission)
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
	members []*app.WorkspaceMemberUpdate,
) error {
	pbMembers, err := a.toWorkspaceMembersPb(members)
	if err != nil {
		return err
	}
	_, err = a.client.UpdateWorkspaceMembers(ctx, &pb.UpdateWorkspaceMembersRequest{
		UserId:      userID,
		WorkspaceId: workspaceID.String(),
		Members:     pbMembers,
	})
	return err
}

func (a *Authorization) GetWorkspaceMembers(ctx context.Context, userID string, workspaceID uuid.UUID) ([]app.AuthorizationWorkspaceMember, error) {
	response, err := a.client.GetWorkspaceMembers(ctx, &pb.GetWorkspaceMembersRequest{
		UserId:      userID,
		WorkspaceId: workspaceID.String(),
	})
	if err != nil {
		return nil, err
	}
	members, err := a.toAppWorkspaceMembers(response.Members)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (a *Authorization) CheckHealth(ctx context.Context) error {
	client := grpc_health_v1.NewHealthClient(a.conn)
	response, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "",
	})
	if err != nil {
		return err
	}
	if response.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return errors.New("authorization service is not healthy")
	}
	return nil
}

func (a *Authorization) toWorkspaceMembersPb(members []*app.WorkspaceMemberUpdate) ([]*pb.WorkspaceMember, error) {
	pbMembers := make([]*pb.WorkspaceMember, 0, len(members))
	for _, member := range members {
		pbMember, err := a.toWorkspaceMemberPb(member)
		if err != nil {
			return nil, err
		}
		pbMembers = append(pbMembers, pbMember)
	}
	return pbMembers, nil
}

func (a *Authorization) toWorkspaceMemberPb(member *app.WorkspaceMemberUpdate) (*pb.WorkspaceMember, error) {
	role, err := a.toWorkspaceRolePb(member.Role)
	if err != nil {
		return nil, err
	}
	return &pb.WorkspaceMember{
		Id:   member.ID,
		Role: role,
	}, nil
}

func (a *Authorization) toAppWorkspaceMembers(members []*pb.WorkspaceMember) ([]app.AuthorizationWorkspaceMember, error) {
	appMembers := make([]app.AuthorizationWorkspaceMember, 0, len(members))
	for _, member := range members {
		appMember, err := a.toAppWorkspaceMember(member)
		if err != nil {
			return nil, err
		}
		appMembers = append(appMembers, appMember)
	}
	return appMembers, nil
}

func (a *Authorization) toAppWorkspaceMember(members *pb.WorkspaceMember) (app.AuthorizationWorkspaceMember, error) {
	role, err := a.toAppWorkspaceRole(members.Role)
	if err != nil {
		return app.AuthorizationWorkspaceMember{}, err
	}
	return app.AuthorizationWorkspaceMember{
		ID:   members.Id,
		Role: role,
	}, nil
}

func (a *Authorization) toAppWorkspaceRole(role pb.WorkspaceRole) (app.WorkspaceRole, error) {
	switch role {
	case pb.WorkspaceRole_WORKSPACE_ROLE_OWNER:
		return app.WorkspaceRoleOwner, nil
	case pb.WorkspaceRole_WORKSPACE_ROLE_EDITOR:
		return app.WorkspaceRoleEditor, nil
	case pb.WorkspaceRole_WORKSPACE_ROLE_VIEWER:
		return app.WorkspaceRoleViewer, nil
	case pb.WorkspaceRole_WORKSPACE_ROLE_UNSPECIFIED:
		return app.WorkspaceRoleUnspecified, nil
	default:
		return app.WorkspaceRoleUnspecified, errs.NewAuthorizationInternal(fmt.Errorf("invalid workspace role: %v", role))
	}
}

func (a *Authorization) toWorkspaceRolePb(role app.WorkspaceRole) (pb.WorkspaceRole, error) {
	switch role {
	case app.WorkspaceRoleOwner:
		return pb.WorkspaceRole_WORKSPACE_ROLE_OWNER, nil
	case app.WorkspaceRoleEditor:
		return pb.WorkspaceRole_WORKSPACE_ROLE_EDITOR, nil
	case app.WorkspaceRoleViewer:
		return pb.WorkspaceRole_WORKSPACE_ROLE_VIEWER, nil
	case app.WorkspaceRoleUnspecified:
		return pb.WorkspaceRole_WORKSPACE_ROLE_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace role: %v", role))
	default:
		return pb.WorkspaceRole_WORKSPACE_ROLE_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace role: %v", role))
	}
}

func (a *Authorization) toAppWorkspacePermission(permission app.WorkspacePermission) (pb.WorkspacePermission, error) {
	switch permission {
	case app.WorkspacePermissionRead:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_READ, nil
	case app.WorkspacePermissionEdit:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_EDIT, nil
	case app.WorkspacePermissionDelete:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_DELETE, nil
	case app.WorkspacePermissionUnspecified:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace permission: %v", permission))
	default:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace permission: %v", permission))
	}
}

func (a *Authorization) toAppWorkspaceItemPermission(permission app.WorkspaceItemPermission) (pb.WorkspaceItemPermission, error) {
	switch permission {
	case app.WorkspaceItemPermissionRead:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_READ, nil
	case app.WorkspaceItemPermissionWrite:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_WRITE, nil
	case app.WorkspaceItemPermissionDelete:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_DELETE, nil
	case app.WorkspaceItemPermissionUnspecified:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace item permission: %v", permission))
	default:
		return pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace item permission: %v", permission))
	}
}
