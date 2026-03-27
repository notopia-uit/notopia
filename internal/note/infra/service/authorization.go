package service

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commongrpc "github.com/notopia-uit/notopia/pkg/common/grpc"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

type Authorization struct {
	client pbconnect.AuthorizationServiceClient
}

var _ app.AuthorizationService = (*Authorization)(nil)

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

func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspacePermission) (bool, errs.Error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceItemPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, errs.Error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceNotePermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, errs.Error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceFolderPermission(ctx context.Context, userID string, workspaceID uuid.UUID, permission app.WorkspaceItemPermission) (bool, errs.Error) {
	return false, errs.NewUnimplemented()
}

func (a *Authorization) CreateWorkspaceWithOwnership(ctx context.Context, userID string, workspaceID uuid.UUID, ownerID uuid.UUID) errs.Error {
	return errs.NewUnimplemented()
}

func (a *Authorization) GetWorkspaceMembers(ctx context.Context, userID string, workspaceID uuid.UUID) ([]*app.WorkspaceMemberInfo, errs.Error) {
	return nil, errs.NewUnimplemented()
}
