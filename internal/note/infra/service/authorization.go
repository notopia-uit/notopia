package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/config"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

type Authorization struct {
	client pbconnect.AuthorizationServiceClient
}

var _ service.Authorization = (*Authorization)(nil)

func NewAuthorization(
	servicesCfg *config.Services,
) *Authorization {
	client := pbconnect.NewAuthorizationServiceClient(
		http.DefaultClient,
		servicesCfg.Authorization.URL,
	)
	return &Authorization{
		client: client,
	}
}

var ProvideAuthorization = NewAuthorization

func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission service.WorkspacePermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceItemPermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission service.WorkspaceItemPermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceNotePermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission service.WorkspaceItemPermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) HasWorkspaceFolderPermission(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, permission service.WorkspaceItemPermission) (bool, error) {
	return false, commonerror.NewUnimplemented()
}

func (a *Authorization) CreateWorkspaceWithOwnership(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, ownerID uuid.UUID) error {
	return commonerror.NewUnimplemented()
}

func (a *Authorization) GetWorkspaceMembers(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]service.WorkspaceMember, error) {
	return nil, commonerror.NewUnimplemented()
}
