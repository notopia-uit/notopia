package authorization

import (
	"context"

	"connectrpc.com/connect"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
	"github.com/notopia-uit/notopia/pkg/pb"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

type GRPCHandler struct{}

var _ pbconnect.AuthorizationServiceHandler = (*GRPCHandler)(nil)

func NewGRPCHandler() *GRPCHandler {
	return &GRPCHandler{}
}

var ProvideGRPCHandler = NewGRPCHandler

func (h *GRPCHandler) CreateWorkspaceWithOwner(context.Context, *connect.Request[pb.CreateWorkspaceWithOwnerRequest]) (*connect.Response[pb.CreateWorkspaceWithOwnerResponse], error) {
	return nil, commonerror.NewUnimplemented("not implemented", "not_implemented", nil)
}

func (h *GRPCHandler) GetWorkspaceMembers(context.Context, *connect.Request[pb.GetWorkspaceMembersRequest]) (*connect.Response[pb.GetWorkspaceMembersResponse], error) {
	return nil, commonerror.NewUnimplemented("not implemented", "not_implemented", nil)
}

func (h *GRPCHandler) HasWorkspacePermission(context.Context, *connect.Request[pb.HasWorkspacePermissionRequest]) (*connect.Response[pb.HasWorkspacePermissionResponse], error) {
	return nil, commonerror.NewUnimplemented("not implemented", "not_implemented", nil)
}

func (h *GRPCHandler) UpdateWorkspaceMembers(context.Context, *connect.Request[pb.UpdateWorkspaceMembersRequest]) (*connect.Response[pb.UpdateWorkspaceMembersResponse], error) {
	return nil, commonerror.NewUnimplemented("not implemented", "not_implemented", nil)
}

func (h *GRPCHandler) HasNotePermission(context.Context, *connect.Request[pb.HasNotePermissionRequest]) (*connect.Response[pb.HasNotePermissionResponse], error) {
	return nil, commonerror.NewUnimplemented("not implemented", "not_implemented", nil)
}

func (h *GRPCHandler) GetUserNotePermissions(context.Context, *connect.Request[pb.GetUserNotePermissionsRequest]) (*connect.Response[pb.GetUserNotePermissionsResponse], error) {
	return nil, commonerror.NewUnimplemented("not implemented", "not_implemented", nil)
}
