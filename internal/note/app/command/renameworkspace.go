package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RenameWorkspace struct {
	ID     uuid.UUID
	UserID string
	Name   string
}

type RenameWorkspaceHandler struct {
	authorization service.Authorization
	workspaceRepo domain.WorkspaceRepo
}

func NewRenameWorkspaceHandler(
	authorization service.Authorization,
	workspaceRepo domain.WorkspaceRepo,
) *RenameWorkspaceHandler {
	return &RenameWorkspaceHandler{
		authorization: authorization,
		workspaceRepo: workspaceRepo,
	}
}

var ProvideRenameWorkspaceHandler = NewRenameWorkspaceHandler

func (h *RenameWorkspaceHandler) Handle(ctx context.Context, cmd *RenameWorkspace) errs.Error {
	hasPermission, err := h.authorization.HasWorkspacePermission(
		ctx,
		cmd.UserID,
		cmd.ID,
		service.WorkspacePermissionEdit,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to rename workspace %s", cmd.UserID, cmd.ID),
		)
	}

	workspace, err := h.workspaceRepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	workspace.Rename(cmd.Name)
	return h.workspaceRepo.Save(ctx, workspace)
}
