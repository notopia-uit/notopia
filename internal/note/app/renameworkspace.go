package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type RenameWorkspace struct {
	ID     uuid.UUID
	UserID string
	Name   string
}

type RenameWorkspaceHandler struct {
	authorizationService AuthorizationService
	workspacerepo        domain.WorkspaceRepo
}

func NewRenameWorkspaceHandler(
	authorizationService AuthorizationService,
	workspacerepo domain.WorkspaceRepo,
) *RenameWorkspaceHandler {
	return &RenameWorkspaceHandler{
		authorizationService: authorizationService,
		workspacerepo:        workspacerepo,
	}
}

var ProvideRenameWorkspaceHandler = NewRenameWorkspaceHandler

func (h *RenameWorkspaceHandler) Handle(ctx context.Context, cmd *RenameWorkspace) errs.Error {
	hasPermission, err := h.authorizationService.HasWorkspacePermission(
		ctx,
		cmd.UserID,
		cmd.ID,
		WorkspacePermissionEdit,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to rename workspace %s", cmd.UserID, cmd.ID),
		)
	}

	workspace, err := h.workspacerepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	workspace.Rename(cmd.Name)
	return h.workspacerepo.Save(ctx, workspace)
}
