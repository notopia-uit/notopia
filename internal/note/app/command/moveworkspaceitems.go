package command

import (
	"context"

	"github.com/google/uuid"
)

type MoveWorkspaceItems struct {
	WorkspaceSlug string
	NoteIDs       []uuid.UUID
	FolderIDs     []uuid.UUID
}

type MoveWorkspaceItemsHandler struct{}

func NewMoveWorkspaceItemsHandler() *MoveWorkspaceItemsHandler {
	return &MoveWorkspaceItemsHandler{}
}

func (h *MoveWorkspaceItemsHandler) Handle(ctx context.Context, cmd *MoveWorkspaceItems) error {
	// TODO: The OpenAPI spec for move-items does not include a destination folder ID.
	// Clarify the spec: does each item carry its own target folderId, or is there a
	// single target? Once clarified, implement:
	// 1. For each noteID: NoteRepo.GetByID + note.MoveToFolder(targetID) + NoteRepo.Save
	// 2. For each folderID: FolderRepo.GetByID + folder.MoveToFolder(targetID) + FolderRepo.Save
	return nil
}
