package authorization

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/internal/authorization/errs"

	"github.com/notopia-uit/notopia/pkg/pb"
)

type GRPCServiceServer struct {
	pb.UnimplementedAuthorizationServiceServer
	app *App
}

func NewGRPCServiceServer(app *App) *GRPCServiceServer {
	return &GRPCServiceServer{
		UnimplementedAuthorizationServiceServer: pb.UnimplementedAuthorizationServiceServer{},
		app:                                     app,
	}
}

var ProvideGRPCServiceServer = NewGRPCServiceServer

var _ pb.AuthorizationServiceServer = (*GRPCServiceServer)(nil)

func (g *GRPCServiceServer) GetUserWorkspaces(ctx context.Context, req *pb.GetUserWorkspacesRequest) (*pb.GetUserWorkspacesResponse, error) {
	workspaces, err := g.app.GetUserWorkspaces.Handle(ctx, app.GetUserWorkspaces{
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}

	pbWorkspaces := make([]*pb.UserWorkspace, len(workspaces))
	for i, w := range workspaces {
		pbWorkspaces[i] = appUserWorkspaceToPb(w)
	}

	return &pb.GetUserWorkspacesResponse{
		Workspaces: pbWorkspaces,
	}, nil
}

func (g *GRPCServiceServer) CreateWorkspaceWithOwner(ctx context.Context, req *pb.CreateWorkspaceWithOwnerRequest) (*pb.CreateWorkspaceWithOwnerResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, err
	}

	if err := g.app.CreateWorkspace.Handle(ctx, app.CreateWorkspace{
		OwnerID:     req.OwnerId,
		WorkspaceID: workspaceID,
	}); err != nil {
		return nil, err
	}

	return &pb.CreateWorkspaceWithOwnerResponse{}, nil
}

func (g *GRPCServiceServer) UpdateWorkspaceMembers(ctx context.Context, req *pb.UpdateWorkspaceMembersRequest) (*pb.UpdateWorkspaceMembersResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, err
	}

	members := make([]app.WorkspaceMember, len(req.Members))
	for i, m := range req.Members {
		members[i] = app.WorkspaceMember{
			ID:   m.Id,
			Role: pbWorkspaceRoleToApp(m.Role),
		}
	}

	if err := g.app.UpdateWorkspaceMembers.Handle(ctx, app.UpdateWorkspaceMembers{
		UserID:      req.UserId,
		WorkspaceID: workspaceID,
		Members:     members,
	}); err != nil {
		return nil, err
	}

	return &pb.UpdateWorkspaceMembersResponse{}, nil
}

func (g *GRPCServiceServer) GetWorkspaceMembers(ctx context.Context, req *pb.GetWorkspaceMembersRequest) (*pb.GetWorkspaceMembersResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, err
	}

	members, err := g.app.GetWorkspaceMembers.Handle(ctx, app.GetWorkspaceMembers{
		UserID:      req.UserId,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, err
	}

	pbMembers := make([]*pb.WorkspaceMember, len(members))
	for i, m := range members {
		pbMembers[i] = appWorkspaceMemberToPb(m)
	}

	return &pb.GetWorkspaceMembersResponse{
		Members: pbMembers,
	}, nil
}

func (g *GRPCServiceServer) HasWorkspacePermission(ctx context.Context, req *pb.HasWorkspacePermissionRequest) (*pb.HasWorkspacePermissionResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, err
	}

	hasPermission, err := g.app.HasWorkspacePermission.Handle(ctx, app.HasWorkspacePermission{
		UserID:      req.MemberId,
		WorkspaceID: workspaceID,
		Permission:  pbWorkspacePermissionToApp(req.Permission),
	})
	if err != nil {
		return nil, err
	}

	return &pb.HasWorkspacePermissionResponse{
		HasPermission: hasPermission,
	}, nil
}

func (g *GRPCServiceServer) HasWorkspaceItemPermission(ctx context.Context, req *pb.HasWorkspaceItemPermissionRequest) (*pb.HasWorkspaceItemPermissionResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	hasPermission, err := g.app.HasWorkspaceItemPermission.Handle(ctx, app.HasWorkspaceItemPermission{
		UserID:      req.MemberId,
		WorkspaceID: workspaceID,
		Permission:  pbWorkspaceItemPermissionToApp(req.Permission),
	})
	if err != nil {
		return nil, err
	}

	return &pb.HasWorkspaceItemPermissionResponse{
		HasPermission: hasPermission,
	}, nil
}

func (g *GRPCServiceServer) GetUserWorkspaceItemPermissions(ctx context.Context, req *pb.GetUserWorkspaceItemPermissionsRequest) (*pb.GetUserWorkspaceItemPermissionsResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	permissions, err := g.app.GetUserWorkspaceItemPermissions.Handle(ctx, app.GetUserWorkspaceItemPermissions{
		UserID:      req.MemberId,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetUserWorkspaceItemPermissionsResponse{
		CanRead:   permissions.Read,
		CanWrite:  permissions.Write,
		CanDelete: permissions.Delete,
	}, nil
}

func (g *GRPCServiceServer) DeleteWorkspace(ctx context.Context, req *pb.DeleteWorkspaceRequest) (*pb.DeleteWorkspaceResponse, error) {
	return nil, errs.Unimplemented
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

func appUserWorkspaceToPb(workspace *app.UserWorkspace) *pb.UserWorkspace {
	return &pb.UserWorkspace{
		Id:   workspace.ID.String(),
		Role: appWorkspaceRoleToPb(workspace.Role),
	}
}

func appWorkspaceMemberToPb(workspace *app.WorkspaceMember) *pb.WorkspaceMember {
	return &pb.WorkspaceMember{
		Id:   workspace.ID,
		Role: appWorkspaceRoleToPb(workspace.Role),
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
		panic("invalid workspace permission")
	default:
		panic("invalid workspace permission")
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
		panic("invalid workspace item permission")
	default:
		panic("invalid workspace item permission")
	}
}
