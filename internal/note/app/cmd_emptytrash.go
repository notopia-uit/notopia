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
	uow              domain.UnitOfWork
}

func NewEmptyTrashHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *EmptyTrashHandler {
	return &EmptyTrashHandler{
		authorizationSvc: authorizationSvc,
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
		folderRepo := r.Folder()
		noteRepo := r.Note()

		trashedFolders, err := folderRepo.GetMany(ctx, &domain.FolderRepoGetManyParams{
			IDs:         nil,
			WorkspaceID: cmd.WorkspaceID,
			TrashOnly:   true,
			TrashedBy:   domain.TrashedByUnspecified,
			ForUpdate:   true,
		})
		if err != nil {
			return err
		}

		for _, folder := range trashedFolders {
			folder.PermanentlyDelete(cmd.UserID)
		}
		if err := folderRepo.SaveMany(ctx, trashedFolders); err != nil {
			return err
		}

		trashedNotes, err := noteRepo.GetMany(ctx, &domain.NoteRepoGetManyParams{
			IDs:         nil,
			WorkspaceID: cmd.WorkspaceID,
			TrashOnly:   true,
			TrashedBy:   domain.TrashedByUnspecified,
			ForUpdate:   true,
		})
		if err != nil {
			return err
		}
		for _, note := range trashedNotes {
			note.PermanentlyDelete(cmd.UserID)
		}
		if err := noteRepo.SaveMany(ctx, trashedNotes); err != nil {
			return err
		}
		return nil
	})
}
