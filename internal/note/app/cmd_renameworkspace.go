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
	uow                  domain.UnitOfWork
}

func NewRenameWorkspaceHandler(
	authorizationService AuthorizationService,
	workspacerepo domain.WorkspaceRepo,
	uow domain.UnitOfWork,
) *RenameWorkspaceHandler {
	return &RenameWorkspaceHandler{
		authorizationService: authorizationService,
		workspacerepo:        workspacerepo,
		uow:                  uow,
	}
}

var ProvideRenameWorkspaceHandler = NewRenameWorkspaceHandler

func (h *RenameWorkspaceHandler) Handle(ctx context.Context, cmd *RenameWorkspace) error {
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

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		workspaceRepo := r.Workspace()
		workspace, err := workspaceRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		workspace.Rename(cmd.Name, cmd.UserID)
		return workspaceRepo.Save(ctx, workspace)
	})
}
