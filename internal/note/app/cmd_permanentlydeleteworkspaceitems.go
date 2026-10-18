package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type PermanentlyDeleteWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type PermanentlyDeleteWorkspaceItemsHandler struct {
	authorizationSvc AuthorizationSvc
	trashService     *domain.TrashService
	uow              domain.UnitOfWork
}

func NewPermanentlyDeleteWorkspaceItemsHandler(
	authorizationSvc AuthorizationSvc,
	trashService *domain.TrashService,
	uow domain.UnitOfWork,
) *PermanentlyDeleteWorkspaceItemsHandler {
	return &PermanentlyDeleteWorkspaceItemsHandler{
		authorizationSvc: authorizationSvc,
		trashService:     trashService,
		uow:              uow,
	}
}

var ProvidePermanentlyDeleteWorkspaceItemsHandler = NewPermanentlyDeleteWorkspaceItemsHandler

type PermanentlyDeleteWorkspaceItemsCmd commonhandler.Cmd[PermanentlyDeleteWorkspaceItems]

var _ PermanentlyDeleteWorkspaceItemsCmd = (*PermanentlyDeleteWorkspaceItemsHandler)(nil)

func (h *PermanentlyDeleteWorkspaceItemsHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteWorkspaceItems) error {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspaceItemPermissionDelete)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to permanently delete items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		targets, err := loadPermanentDeleteTargets(ctx, r, cmd.NoteIDs, cmd.FolderIDs)
		if err != nil {
			return err
		}
		targets.delete(h.trashService, cmd.UserID)
		return savePermanentDeleteTargets(ctx, r, targets)
	})
}
