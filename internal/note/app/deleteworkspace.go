package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type DeleteWorkspace struct {
	ID     uuid.UUID
	UserID string
}

type DeleteWorkspaceHandler struct {
	authorizationService AuthorizationService
	workspaceRepo        domain.WorkspaceRepo
}

func NewDeleteWorkspaceHandler(
	authorizationService AuthorizationService,
	workspaceRepo domain.WorkspaceRepo,
) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{
		authorizationService: authorizationService,
		workspaceRepo:        workspaceRepo,
	}
}

var ProvideDeleteWorkspaceHandler = NewDeleteWorkspaceHandler

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd *DeleteWorkspace) error {
	hasPermission, err := h.authorizationService.HasWorkspacePermission(
		ctx,
		cmd.UserID,
		cmd.ID,
		WorkspacePermissionDelete,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return newErrDeleteWorkspaceForbidden(cmd.UserID, cmd.ID)
	}

	workspace, err := h.workspaceRepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	workspace.Delete()
	return h.workspaceRepo.Save(ctx, workspace)
}

var ErrCodeDeleteWorkspaceForbidden = "DeleteWorkspace_1"

func newErrDeleteWorkspaceForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to delete workspace %q", userID, workspaceID.String()),
		ErrCodeDeleteWorkspaceForbidden,
		nil,
	)
}
