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

func (h *GRPCHandler) CreateWorkspace(context.Context, *connect.Request[pb.CreateWorkspaceRequest]) (*connect.Response[pb.CreateWorkspaceResponse], error) {
	return nil, commonerror.NewUnimplemented()
}

func (h *GRPCHandler) UpdateWorkspaceMembers(context.Context, *connect.Request[pb.UpdateWorkspaceMembersRequest]) (*connect.Response[pb.UpdateWorkspaceMembersResponse], error) {
	return nil, commonerror.NewUnimplemented()
}

func (h *GRPCHandler) GetWorkspaceMembers(context.Context, *connect.Request[pb.GetWorkspaceMembersRequest]) (*connect.Response[pb.GetWorkspaceMembersResponse], error) {
	return nil, commonerror.NewUnimplemented()
}

func (h *GRPCHandler) HasWorkspacePermission(context.Context, *connect.Request[pb.HasWorkspacePermissionRequest]) (*connect.Response[pb.HasWorkspacePermissionResponse], error) {
	return nil, commonerror.NewUnimplemented()
}

func (h *GRPCHandler) HasWorkspaceItemPermission(context.Context, *connect.Request[pb.HasWorkspaceItemPermissionRequest]) (*connect.Response[pb.HasWorkspaceItemPermissionResponse], error) {
	return nil, commonerror.NewUnimplemented()
}

func (h *GRPCHandler) GetUserWorkspaceItemPermissions(context.Context, *connect.Request[pb.GetUserWorkspaceItemPermissionsRequest]) (*connect.Response[pb.GetUserWorkspaceItemPermissionsResponse], error) {
	return nil, commonerror.NewUnimplemented()
}
