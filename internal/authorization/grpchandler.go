package authorization

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"

	"github.com/notopia-uit/notopia/pkg/pb"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

type GRPCHandler struct {
	app *App
}

var _ pbconnect.AuthorizationServiceHandler = (*GRPCHandler)(nil)

func NewGRPCHandler(app *App) *GRPCHandler {
	return &GRPCHandler{app: app}
}

var ProvideGRPCHandler = NewGRPCHandler

func (h *GRPCHandler) CreateWorkspace(ctx context.Context, req *connect.Request[pb.CreateWorkspaceRequest]) (*connect.Response[pb.CreateWorkspaceResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, err
	}

	if err := h.app.CreateWorkspace.Handle(app.CreateWorkspace{
		UserID:      req.Msg.UserId,
		WorkspaceID: workspaceID,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.CreateWorkspaceResponse{}), nil
}

func (h *GRPCHandler) UpdateWorkspaceMembers(ctx context.Context, req *connect.Request[pb.UpdateWorkspaceMembersRequest]) (*connect.Response[pb.UpdateWorkspaceMembersResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, err
	}

	members := make([]app.WorkspaceMember, len(req.Msg.Members))
	for i, m := range req.Msg.Members {
		members[i] = app.WorkspaceMember{
			ID:   m.Id,
			Role: pbWorkspaceRoleToApp(m.Role),
		}
	}

	if err := h.app.UpdateWorkspaceMembers.Handle(ctx, app.UpdateWorkspaceMembers{
		UserID:      req.Msg.UserId,
		WorkspaceID: workspaceID,
		Members:     members,
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.UpdateWorkspaceMembersResponse{}), nil
}

func (h *GRPCHandler) GetWorkspaceMembers(ctx context.Context, req *connect.Request[pb.GetWorkspaceMembersRequest]) (*connect.Response[pb.GetWorkspaceMembersResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, err
	}

	members, err := h.app.GetWorkspaceMembers.Handle(app.GetWorkspaceMembers{
		UserID:      req.Msg.UserId,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, err
	}

	pbMembers := make([]*pb.WorkspaceMember, len(members))
	for i, m := range members {
		pbMembers[i] = &pb.WorkspaceMember{
			Id:   m.ID,
			Role: appWorkspaceRoleToPb(m.Role),
		}
	}

	return connect.NewResponse(&pb.GetWorkspaceMembersResponse{
		Members: pbMembers,
	}), nil
}

func (h *GRPCHandler) HasWorkspacePermission(ctx context.Context, req *connect.Request[pb.HasWorkspacePermissionRequest]) (*connect.Response[pb.HasWorkspacePermissionResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, err
	}

	hasPermission, err := h.app.HasWorkspacePermission.Handle(app.HasWorkspacePermission{
		UserID:      req.Msg.MemberId,
		WorkspaceID: workspaceID,
		Permission:  pbWorkspacePermissionToApp(req.Msg.Permission),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.HasWorkspacePermissionResponse{
		HasPermission: hasPermission,
	}), nil
}

func (h *GRPCHandler) HasWorkspaceItemPermission(ctx context.Context, req *connect.Request[pb.HasWorkspaceItemPermissionRequest]) (*connect.Response[pb.HasWorkspaceItemPermissionResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, err
	}
	hasPermission, err := h.app.HasWorkspaceItemPermission.Handle(app.HasWorkspaceItemPermission{
		UserID:      req.Msg.MemberId,
		WorkspaceID: workspaceID,
		Permission:  pbWorkspaceItemPermissionToApp(req.Msg.Permission),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.HasWorkspaceItemPermissionResponse{
		HasPermission: hasPermission,
	}), nil
}

func (h *GRPCHandler) GetUserWorkspaceItemPermissions(ctx context.Context, req *connect.Request[pb.GetUserWorkspaceItemPermissionsRequest]) (*connect.Response[pb.GetUserWorkspaceItemPermissionsResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, err
	}
	permissions, err := h.app.GetUserWorkspaceItemPermissions.Handle(app.GetUserWorkspaceItemPermissions{
		UserID:      req.Msg.MemberId,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetUserWorkspaceItemPermissionsResponse{
		CanRead:   permissions.Read,
		CanWrite:  permissions.Write,
		CanDelete: permissions.Delete,
	}), nil
}

func pbWorkspaceRoleToApp(role pb.WorkspaceRole) app.WorkspaceRole {
	switch role {
	case pb.WorkspaceRole_WORKSPACE_ROLE_OWNER:
		return app.WorkspaceRoleOwner
	case pb.WorkspaceRole_WORKSPACE_ROLE_EDITOR:
		return app.WorkspaceRoleEditor
	case pb.WorkspaceRole_WORKSPACE_ROLE_VIEWER:
		return app.WorkspaceRoleViewer
	case pb.WorkspaceRole_WORKSPACE_ROLE_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}

func appWorkspaceRoleToPb(role app.WorkspaceRole) pb.WorkspaceRole {
	switch role {
	case app.WorkspaceRoleOwner:
		return pb.WorkspaceRole_WORKSPACE_ROLE_OWNER
	case app.WorkspaceRoleEditor:
		return pb.WorkspaceRole_WORKSPACE_ROLE_EDITOR
	case app.WorkspaceRoleViewer:
		return pb.WorkspaceRole_WORKSPACE_ROLE_VIEWER
	default:
		return pb.WorkspaceRole_WORKSPACE_ROLE_UNSPECIFIED
	}
}

func pbWorkspacePermissionToApp(perm pb.WorkspacePermission) app.WorkspacePermission {
	switch perm {
	case pb.WorkspacePermission_WORKSPACE_PERMISSION_READ:
		return app.WorkspacePermissionRead
	case pb.WorkspacePermission_WORKSPACE_PERMISSION_EDIT:
		return app.WorkspacePermissionEdit
	case pb.WorkspacePermission_WORKSPACE_PERMISSION_DELETE:
		return app.WorkspacePermissionDelete
	case pb.WorkspacePermission_WORKSPACE_PERMISSION_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}

func pbWorkspaceItemPermissionToApp(perm pb.WorkspaceItemPermission) app.WorkspaceItemPermission {
	switch perm {
	case pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_READ:
		return app.WorkspaceItemPermissionRead
	case pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_WRITE:
		return app.WorkspaceItemPermissionWrite
	case pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_DELETE:
		return app.WorkspaceItemPermissionDelete
	case pb.WorkspaceItemPermission_WORKSPACE_ITEM_PERMISSION_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}
