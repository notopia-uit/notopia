package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	folderRepo           domain.FolderRepo
	uow                  domain.UnitOfWork
	workspaceEventPubSub WorkspaceEventPubSub
}

func NewMoveWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	uow domain.UnitOfWork,
	workspaceEventPubSub WorkspaceEventPubSub,
) *MoveWorkspaceItemsHandler {
	return &MoveWorkspaceItemsHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		folderRepo:           folderRepo,
		uow:                  uow,
		workspaceEventPubSub: workspaceEventPubSub,
	}
}

var ProvideMoveWorkspaceItemsHandler = NewMoveWorkspaceItemsHandler

func (h *MoveWorkspaceItemsHandler) Handle(ctx context.Context, cmd *MoveWorkspaceItems) errs.Error {
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		WorkspaceItemPermissionWrite,
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
			domain.NewFolderRepoGetManyParamsByIDs(cmd.FolderIDs).
				WithWorkspaceID(cmd.WorkspaceID).
				WithForUpdate())
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
			domain.NewNoteRepoGetManyParamsByIDs(cmd.NoteIDs).
				WithForUpdate(),
		)
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

	return h.workspaceEventPubSub.Publish(ctx, cmd.WorkspaceID, cmd.UserID, workspaceEvents...)
}
