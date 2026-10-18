package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type EmptyTrash struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type EmptyTrashHandler struct {
	authorizationSvc AuthorizationSvc
	trashService     *domain.TrashService
	uow              domain.UnitOfWork
}

func NewEmptyTrashHandler(
	authorizationSvc AuthorizationSvc,
	trashService *domain.TrashService,
	uow domain.UnitOfWork,
) *EmptyTrashHandler {
	return &EmptyTrashHandler{
		authorizationSvc: authorizationSvc,
		trashService:     trashService,
		uow:              uow,
	}
}

var ProvideEmptyTrashHandler = NewEmptyTrashHandler

type EmptyTrashCmd commonhandler.Cmd[EmptyTrash]

var _ EmptyTrashCmd = (*EmptyTrashHandler)(nil)

func (h *EmptyTrashHandler) Handle(ctx context.Context, cmd *EmptyTrash) error {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspaceItemPermissionDelete)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %q does not have permission to delete workspace item in workspace %q", cmd.UserID, cmd.WorkspaceID),
		)
	}

	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		trashedNotes, err := r.Note().GetMany(ctx, &domain.NoteRepoGetManyParams{
			IDs:         nil,
			WorkspaceID: cmd.WorkspaceID,
			TrashOnly:   true,
			TrashedBy:   domain.TrashedByUnspecified,
			ForUpdate:   true,
		})
		if err != nil {
			return err
		}
		trashedFolders, err := r.Folder().GetMany(ctx, &domain.FolderRepoGetManyParams{
			IDs:         nil,
			WorkspaceID: cmd.WorkspaceID,
			TrashOnly:   true,
			TrashedBy:   domain.TrashedByUnspecified,
			ForUpdate:   true,
		})
		if err != nil {
			return err
		}

		noteIDs := make([]uuid.UUID, 0, len(trashedNotes))
		for _, n := range trashedNotes {
			noteIDs = append(noteIDs, n.ID())
		}
		folderIDs := make([]uuid.UUID, 0, len(trashedFolders))
		for _, f := range trashedFolders {
			folderIDs = append(folderIDs, f.ID())
		}

		targets, err := loadPermanentDeleteTargets(ctx, r, noteIDs, folderIDs)
		if err != nil {
			return err
		}
		targets.delete(h.trashService, cmd.UserID)
		return savePermanentDeleteTargets(ctx, r, targets)
	})
}
