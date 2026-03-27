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

type MoveWorkspaceItems struct {
	UserID              string
	WorkspaceID         uuid.UUID
	NoteIDs             []uuid.UUID
	FolderIDs           []uuid.UUID
	DestinationFolderID uuid.UUID
}

type MoveWorkspaceItemsHandler struct {
	authorization service.Authorization
	noteRepo      domain.NoteRepo
	folderRepo    domain.FolderRepo
	uow           domain.UnitOfWork
	eventPubSub   pubsub.WorkspaceEvent
}

func NewMoveWorkspaceItemsHandler(
	authorization service.Authorization,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	uow domain.UnitOfWork,
	eventPubSub pubsub.WorkspaceEvent,
) *MoveWorkspaceItemsHandler {
	return &MoveWorkspaceItemsHandler{
		authorization: authorization,
		noteRepo:      noteRepo,
		folderRepo:    folderRepo,
		uow:           uow,
		eventPubSub:   eventPubSub,
	}
}

var ProvideMoveWorkspaceItemsHandler = NewMoveWorkspaceItemsHandler

func (h *MoveWorkspaceItemsHandler) Handle(ctx context.Context, cmd *MoveWorkspaceItems) errs.Error {
	hasPermission, err := h.authorization.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		service.WorkspaceItemPermissionWrite,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to move items in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	folderValid, err := h.folderRepo.AreAllInWorkspace(ctx, cmd.FolderIDs, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if !folderValid {
		return errs.NewInvalid(
			fmt.Sprintf("one or more folders do not belong to workspace %s", cmd.WorkspaceID),
		)
	}

	noteValid, err := h.noteRepo.AreAllInWorkspace(ctx, cmd.NoteIDs, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if !noteValid {
		return errs.NewInvalid(
			fmt.Sprintf("one or more notes do not belong to workspace %s", cmd.WorkspaceID),
		)
	}

	destinationFolder, err := h.folderRepo.GetByID(ctx, cmd.DestinationFolderID, false)
	if err != nil {
		return err
	}

	if destinationFolder.WorkspaceID() != cmd.WorkspaceID {
		return errs.NewInvalid(
			fmt.Sprintf("destination folder %s does not belong to workspace %s", cmd.DestinationFolderID, cmd.WorkspaceID),
		)
	}

	var folders []*domain.Folder
	var notes []*domain.Note

	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) errs.Error {
		folderRepo := r.Folder()
		noteRepo := r.Note()
		folders, err = folderRepo.GetMany(ctx,
			&domain.FolderRepoGetManyParams{
				IDs:       cmd.FolderIDs,
				ForUpdate: true,
			})
		if err != nil {
			return err
		}
		for _, folder := range folders {
			folder.MoveToFolder(cmd.DestinationFolderID)
		}
		if err := folderRepo.SaveMany(ctx, folders); err != nil {
			return err
		}
		notes, err = noteRepo.GetMany(ctx,
			&domain.NoteRepoGetManyParams{
				IDs:       cmd.NoteIDs,
				ForUpdate: true,
			})
		if err != nil {
			return err
		}
		for _, note := range notes {
			note.MoveToFolder(cmd.DestinationFolderID)
		}
		if err := noteRepo.SaveMany(ctx, notes); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	workspaceEvents := make([]domain.Event, 0, len(cmd.FolderIDs)+len(cmd.NoteIDs))
	for _, folder := range folders {
		workspaceEvents = append(workspaceEvents, folder.PopEvents()...)
	}
	for _, note := range notes {
		workspaceEvents = append(workspaceEvents, note.PopEvents()...)
	}

	return h.eventPubSub.Publish(ctx, cmd.WorkspaceID, cmd.UserID, workspaceEvents...)
}
