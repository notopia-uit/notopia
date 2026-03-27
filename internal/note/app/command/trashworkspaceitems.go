package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type TrashWorkspaceItems struct {
	WorkspaceID uuid.UUID
	UserID      string
	NoteIDs     []uuid.UUID
	FolderIDs   []uuid.UUID
}

type TrashWorkspaceItemsHandler struct {
	authorization service.Authorization
	uow           domain.UnitOfWork
	trashService  *domain.TrashService
	eventPubSub   pubsub.WorkspaceEvent
}

func NewTrashWorkspaceItemsHandler(
	authorization service.Authorization,
	uow domain.UnitOfWork,
	trashService *domain.TrashService,
	eventPubSub pubsub.WorkspaceEvent,
) *TrashWorkspaceItemsHandler {
	return &TrashWorkspaceItemsHandler{
		authorization: authorization,
		uow:           uow,
		trashService:  trashService,
		eventPubSub:   eventPubSub,
	}
}

var ProvideTrashWorkspaceItemsHandler = NewTrashWorkspaceItemsHandler

func (h *TrashWorkspaceItemsHandler) Handle(ctx context.Context, cmd *TrashWorkspaceItems) errs.Error {
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		service.WorkspaceItemPermissionDelete,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to trash items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	var workspaceEvents []domain.Event

	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) errs.Error {
		noteRepo := r.Note()
		folderRepo := r.Folder()

		if len(cmd.NoteIDs) == 0 && len(cmd.FolderIDs) == 0 {
			return nil
		}

		workspaceNotes, err := noteRepo.GetMany(ctx,
			&domain.NoteRepoGetManyParams{
				WorkspaceID: &cmd.WorkspaceID,
			})
		if err != nil {
			return err
		}

		workspaceFolders, err := folderRepo.GetMany(ctx,
			&domain.FolderRepoGetManyParams{
				WorkspaceID: &cmd.WorkspaceID,
			})
		if err != nil {
			return err
		}

		workspaceNotePtrs := workspaceNotes
		workspaceFolderPtrs := workspaceFolders

		var notes []*domain.Note
		if len(cmd.NoteIDs) > 0 {
			notes, err = noteRepo.GetMany(ctx,
				&domain.NoteRepoGetManyParams{
					IDs:       cmd.NoteIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}

			if err := h.trashService.TrashNotes(notes); err != nil {
				return err
			}
		}

		var folders []*domain.Folder
		if len(cmd.FolderIDs) > 0 {
			folders, err = folderRepo.GetMany(ctx,
				&domain.FolderRepoGetManyParams{
					IDs:       cmd.FolderIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}

			if err := h.trashService.TrashFolders(&workspaceNotePtrs, &workspaceFolderPtrs, folders); err != nil {
				return err
			}
		}

		if len(workspaceNotePtrs) > 0 {
			if err := noteRepo.SaveMany(ctx, workspaceNotePtrs); err != nil {
				return err
			}
			for _, note := range workspaceNotePtrs {
				workspaceEvents = append(workspaceEvents, note.PopEvents()...)
			}
		}

		if len(workspaceFolderPtrs) > 0 {
			if err := folderRepo.SaveMany(ctx, workspaceFolderPtrs); err != nil {
				return err
			}
			for _, folder := range workspaceFolderPtrs {
				workspaceEvents = append(workspaceEvents, folder.PopEvents()...)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(workspaceEvents) > 0 {
		return h.eventPubSub.Publish(ctx, cmd.WorkspaceID, cmd.UserID, workspaceEvents...)
	}

	return nil
}
