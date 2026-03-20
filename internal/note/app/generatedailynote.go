package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type GenerateDailyNote struct {
	NoteID      uuid.UUID
	WorkspaceID uuid.UUID
}

type GenerateDailyNoteHandler struct {
	noteRepo      domain.NoteRepo
	folderRepo    domain.FolderRepo
	workspaceRepo domain.WorkspaceRepo
}

func NewGenerateDailyNoteHandler(
	noteRepo domain.NoteRepo,
	folderRepo domain.FolderRepo,
	workspaceRepo domain.WorkspaceRepo,
) *GenerateDailyNoteHandler {
	return &GenerateDailyNoteHandler{
		noteRepo:      noteRepo,
		folderRepo:    folderRepo,
		workspaceRepo: workspaceRepo,
	}
}

var ProvideGenerateDailyNoteHandler = NewGenerateDailyNoteHandler

func (h *GenerateDailyNoteHandler) Handle(ctx context.Context, cmd *GenerateDailyNote) (*uuid.UUID, error) {
	// WARN: Unimplemented stub - returns nil, nil without any logic.
	// TODO: No domain method for generating a daily note. Implement logic to:
	// 1. Find or create a "Daily Notes" folder in the workspace root
	//    - Use folderRepo.GetByWorkspaceID() to find existing, or NewFolder if not found
	// 2. Find or create today's note in that folder (named e.g. "2026-03-20")
	//    - Compare note names to find today's date, or use createdAt timestamp
	// 3. Return the note ID for the Content-Location response header
	// Expected HTTP response: 201 Created with Content-Location header pointing to the new note
	// Consider:
	//   - Using time.Now().Format("2006-01-02") for note naming
	//   - Publishing NoteCreatedEvent if new note is created
	//   - Returning cmd.NoteID if pre-generated, or generated ID if created
	return nil, nil
}
