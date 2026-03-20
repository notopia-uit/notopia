package service

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
	commongrpc "github.com/notopia-uit/notopia/pkg/common/grpc"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

type Authorization struct {
	client pbconnect.AuthorizationServiceClient
}

var _ app.Authorization = (*Authorization)(nil)

func NewAuthorization(
	servicesCfg *config.Services,
) *Authorization {
	client := pbconnect.NewAuthorizationServiceClient(
		http.DefaultClient,
		servicesCfg.Authorization.URL,
		connect.WithInterceptors(
			commongrpc.NewClientErrorInterceptor(),
		),
	)
	return &Authorization{
		client: client,
	}
}

var ProvideAuthorization = NewAuthorization

func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspacePermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceItemPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceNotePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceFolderPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) CreateWorkspaceWithOwnership(ctx context.Context, userID string, workspaceID uuid.UUID, ownerID uuid.UUID) error {
	return commonerror.NewUnimplemented()
}

func (a *Authorization) GetWorkspaceMembers(ctx context.Context, userID string, workspaceID uuid.UUID) ([]app.WorkspaceMemberInfo, error) {
	return nil, commonerror.NewUnimplemented()
}
