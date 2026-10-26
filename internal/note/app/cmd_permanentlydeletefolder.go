package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type PermanentlyDeleteFolder struct {
	ID     uuid.UUID
	UserID string
}

type PermanentlyDeleteFolderHandler struct {
	authorizationSvc AuthorizationSvc
	trashService     *domain.TrashService
	uow              domain.UnitOfWork
}

func NewPermanentlyDeleteFolderHandler(
	authorizationSvc AuthorizationSvc,
	trashService *domain.TrashService,
	uow domain.UnitOfWork,
) *PermanentlyDeleteFolderHandler {
	return &PermanentlyDeleteFolderHandler{
		authorizationSvc: authorizationSvc,
		trashService:     trashService,
		uow:              uow,
	}
}

var ProvidePermanentlyDeleteFolderHandler = NewPermanentlyDeleteFolderHandler

type PermanentlyDeleteFolderCmd commonhandler.Cmd[PermanentlyDeleteFolder]

var _ PermanentlyDeleteFolderCmd = (*PermanentlyDeleteFolderHandler)(nil)

func (h *PermanentlyDeleteFolderHandler) Handle(ctx context.Context, cmd *PermanentlyDeleteFolder) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		workspaceID, err := folderRepo.GetWorkspaceIDByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, workspaceID, WorkspaceItemPermissionDelete)
		if err != nil {
			return err
		}
		if !hasPermission {
			return errs.NewForbidden(
				fmt.Sprintf("user %s does not have permission to delete folder %s", cmd.UserID, cmd.ID),
			)
		}
		targets, err := loadPermanentDeleteTargets(ctx, r, nil, []uuid.UUID{cmd.ID})
		if err != nil {
			return err
		}
		targets.delete(h.trashService, cmd.UserID)
		return savePermanentDeleteTargets(ctx, r, targets)
	})
}
