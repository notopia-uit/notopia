package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
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

func (h *GenerateDailyNoteHandler) Handle(ctx context.Context, cmd *GenerateDailyNote) (*uuid.UUID, errs.Error) {
	return nil, nil
}
