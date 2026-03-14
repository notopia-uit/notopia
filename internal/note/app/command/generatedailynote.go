package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type GenerateDailyNote struct {
	WorkspaceID uuid.UUID
}

type GenerateDailyNoteHandler struct {
	noterepo      domain.NoteRepo
	folderrepo    domain.FolderRepo
	workspacerepo domain.WorkspaceRepo
}

func NewGenerateDailyNoteHandler(
	noterepo domain.NoteRepo,
	folderrepo domain.FolderRepo,
	workspacerepo domain.WorkspaceRepo,
) *GenerateDailyNoteHandler {
	return &GenerateDailyNoteHandler{
		noterepo:      noterepo,
		folderrepo:    folderrepo,
		workspacerepo: workspacerepo,
	}
}

func (h *GenerateDailyNoteHandler) Handle(ctx context.Context, cmd *GenerateDailyNote) (*uuid.UUID, error) {
	// TODO: No domain method for generating a daily note. Implement logic to:
	// 1. Find or create a "Daily Notes" folder in the workspace root
	// 2. Find or create today's note (named e.g. "2026-03-14") in that folder
	// Return the note ID for the Content-Location response header.
	return nil, nil
}
