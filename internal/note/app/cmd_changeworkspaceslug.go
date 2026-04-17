package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type ChangeWorkspaceSlug struct {
	ID     uuid.UUID
	Slug   string
	UserID string
}

type ChangeWorkspaceSlugHandler struct {
	authorizationSvc AuthorizationSvc
	uow                  domain.UnitOfWork
}

func NewChangeWorkspaceSlugHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *ChangeWorkspaceSlugHandler {
	return &ChangeWorkspaceSlugHandler{
		authorizationSvc: authorizationSvc,
		uow:                  uow,
	}
}

var ProvideChangeWorkspaceSlugHandler = NewChangeWorkspaceSlugHandler

func (h *ChangeWorkspaceSlugHandler) Handle(ctx context.Context, cmd *ChangeWorkspaceSlug) error {
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(ctx, cmd.UserID, cmd.ID, WorkspacePermissionEdit)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to edit workspace %s", cmd.UserID, cmd.ID),
		)
	}

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		workspaceRepo := r.Workspace()
		workspace, err := workspaceRepo.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return err
		}
		if err := workspace.ChangeSlug(cmd.Slug, cmd.UserID); err != nil {
			return err
		}
		return workspaceRepo.Save(ctx, workspace)
	})
}
