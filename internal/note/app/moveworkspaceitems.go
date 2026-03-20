package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
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
	workspaceEvent       WorkspaceEventPubSub
}

func NewMoveWorkspaceItemsHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	uow domain.UnitOfWork,
	workspaceEvent WorkspaceEventPubSub,
) *MoveWorkspaceItemsHandler {
	return &MoveWorkspaceItemsHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		folderRepo:           folderRepo,
		uow:                  uow,
		workspaceEvent:       workspaceEvent,
	}
}

var ProvideMoveWorkspaceItemsHandler = NewMoveWorkspaceItemsHandler

func (h *MoveWorkspaceItemsHandler) Handle(ctx context.Context, cmd *MoveWorkspaceItems) error {
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
		return newErrMoveWorkspaceItemsForbidden(cmd.UserID, cmd.WorkspaceID)
	}

	folderValid, err := h.folderRepo.AreAllInWorkspace(ctx, cmd.FolderIDs, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if !folderValid {
		return newErrMoveWorkspaceItemsInvalidFolder(cmd.WorkspaceID)
	}

	noteValid, err := h.noteRepo.AreAllInWorkspace(ctx, cmd.NoteIDs, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if !noteValid {
		return newErrMoveWorkspaceItemsInvalidNote(cmd.WorkspaceID)
	}

	destinationFolder, err := h.folderRepo.GetByID(ctx, cmd.DestinationFolderID, false)
	if err != nil {
		return err
	}

	if destinationFolder.WorkspaceID() != cmd.WorkspaceID {
		return newErrMoveWorkspaceItemsInvalidDestination(cmd.WorkspaceID)
	}

	var folders []domain.Folder
	var notes []domain.Note

	err = h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		noteRepo := r.Note()
		folders, err = folderRepo.GetByIDs(ctx, cmd.FolderIDs, true)
		if err != nil {
			return err
		}
		for _, folder := range folders {
			folder.MoveToFolder(cmd.DestinationFolderID)
		}
		if err := folderRepo.SaveMany(ctx, folders); err != nil {
			return err
		}
		notes, err = noteRepo.GetByIDs(ctx, cmd.NoteIDs, true)
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

	err = h.workspaceEvent.Publish(ctx, cmd.WorkspaceID, cmd.UserID, workspaceEvents...)
	if err != nil {
		return err
	}
	return nil
}

var (
	ErrCodeMoveWorkspaceItemsForbidden          = "MoveWorkspaceItems_1"
	ErrCodeMoveWorkspaceItemsInvalidFolder      = "MoveWorkspaceItems_2"
	ErrCodeMoveWorkspaceItemsInvalidNote        = "MoveWorkspaceItems_3"
	ErrCodeMoveWorkspaceItemsInvalidDestination = "MoveWorkspaceItems_4"
)

func newErrMoveWorkspaceItemsForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("User %q does not have permission to move items in workspace %q", userID, workspaceID.String()),
		ErrCodeMoveWorkspaceItemsForbidden,
		nil,
	)
}

func newErrMoveWorkspaceItemsInvalidFolder(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInvalid(
		fmt.Sprintf("One or more folders are not in workspace %q", workspaceID.String()),
		ErrCodeMoveWorkspaceItemsInvalidFolder,
		nil,
	)
}

func newErrMoveWorkspaceItemsInvalidNote(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInvalid(
		fmt.Sprintf("One or more notes are not in workspace %q", workspaceID.String()),
		ErrCodeMoveWorkspaceItemsInvalidNote,
		nil,
	)
}

func newErrMoveWorkspaceItemsInvalidDestination(workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInvalid(
		fmt.Sprintf("Destination folder is not in workspace %q", workspaceID.String()),
		ErrCodeMoveWorkspaceItemsInvalidDestination,
		nil,
	)
}
