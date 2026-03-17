package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/service"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type MoveWorkspaceItems struct {
	UserID              uuid.UUID
	WorkspaceID         uuid.UUID
	NoteIDs             []uuid.UUID
	FolderIDs           []uuid.UUID
	DestinationFolderID uuid.UUID
}

type MoveWorkspaceItemsHandler struct {
	authorizationService service.Authorization
	noteRepo             domain.NoteRepo
	folderRepo           domain.FolderRepo
}

func (h *MoveWorkspaceItemsHandler) Handle(ctx context.Context, cmd *MoveWorkspaceItems) error {
	// TODO: The OpenAPI spec for move-items does not include a destination folder ID.
	// Clarify the spec: does each item carry its own target folderId, or is there a
	// single target? Once clarified, implement:
	// 1. For each noteID: NoteRepo.GetByID + note.MoveToFolder(targetID) + NoteRepo.Save
	// 2. For each folderID: FolderRepo.GetByID + folder.MoveToFolder(targetID) + FolderRepo.Save

	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		cmd.UserID,
		cmd.WorkspaceID,
		service.WorkspaceItemPermissionWrite,
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

	destinationFolder, err := h.folderRepo.GetByID(ctx, cmd.DestinationFolderID)
	if err != nil {
		return err
	}

	if destinationFolder.WorkspaceID() != cmd.WorkspaceID {
		return newErrMoveWorkspaceItemsInvalidDestination(cmd.WorkspaceID)
	}
	// TODO: continue
	return nil
}

var (
	ErrCodeMoveWorkspaceItemsForbidden          = "MoveWorkspaceItems_1"
	ErrCodeMoveWorkspaceItemsInvalidFolder      = "MoveWorkspaceItems_2"
	ErrCodeMoveWorkspaceItemsInvalidNote        = "MoveWorkspaceItems_3"
	ErrCodeMoveWorkspaceItemsInvalidDestination = "MoveWorkspaceItems_4"
)

func newErrMoveWorkspaceItemsForbidden(userID uuid.UUID, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("User %q does not have permission to move items in workspace %q", userID.String(), workspaceID.String()),
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
