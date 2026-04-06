package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
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
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to delete workspace %s", cmd.UserID, cmd.ID),
		)
	}

	workspace, err := h.workspaceRepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	workspace.Delete(cmd.UserID)
	return h.workspaceRepo.Save(ctx, workspace)
}
