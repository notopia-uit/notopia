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
	uow              domain.UnitOfWork
}

func NewPermanentlyDeleteWorkspaceItemsHandler(
	authorizationSvc AuthorizationSvc,
	uow domain.UnitOfWork,
) *PermanentlyDeleteWorkspaceItemsHandler {
	return &PermanentlyDeleteWorkspaceItemsHandler{
		authorizationSvc: authorizationSvc,
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
		folderRepo := r.Folder()
		noteRepo := r.Note()

		notesByID := map[uuid.UUID]*domain.Note{}
		if len(cmd.NoteIDs) > 0 {
			notes, err := noteRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.NoteRepoGetManyParams{
					IDs:       cmd.NoteIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}
			for _, n := range notes {
				notesByID[n.ID()] = n
			}
		}

		var folders []*domain.Folder
		if len(cmd.FolderIDs) > 0 {
			var err error
			folders, err = folderRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.FolderRepoGetManyParams{
					IDs:       cmd.FolderIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}
			for _, folder := range folders {
				children, err := noteRepo.GetRecursiveChildrenFromFolder(ctx, folder.ID(), true)
				if err != nil {
					return err
				}
				for _, n := range children {
					notesByID[n.ID()] = n
				}
			}
		}

		notes := make([]*domain.Note, 0, len(notesByID))
		for _, n := range notesByID {
			n.PermanentlyDelete(cmd.UserID)
			notes = append(notes, n)
		}
		if err := noteRepo.SaveMany(ctx, notes); err != nil {
			return err
		}

		for _, folder := range folders {
			folder.PermanentlyDelete(cmd.UserID)
		}
		return folderRepo.SaveMany(ctx, folders)
	})
}
