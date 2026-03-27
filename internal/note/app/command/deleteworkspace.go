package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type DeleteWorkspace struct {
	ID     uuid.UUID
	UserID string
}

type DeleteWorkspaceHandler struct {
	authorization service.Authorization
	workspaceRepo domain.WorkspaceRepo
}

func NewDeleteWorkspaceHandler(
	authorization service.Authorization,
	workspaceRepo domain.WorkspaceRepo,
) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{
		authorization: authorization,
		workspaceRepo: workspaceRepo,
	}
}

var ProvideDeleteWorkspaceHandler = NewDeleteWorkspaceHandler

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd *DeleteWorkspace) errs.Error {
	hasPermission, err := h.authorization.HasWorkspacePermission(
		ctx,
		cmd.UserID,
		cmd.ID,
		service.WorkspacePermissionDelete,
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
	workspace.Delete()
	return h.workspaceRepo.Save(ctx, workspace)
}
