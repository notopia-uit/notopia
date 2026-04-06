package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetNote struct {
	ID             uuid.UUID
	ExcludeTrashed bool

	UserID string
}

type GetNoteReadModel interface {
	GetNote(ctx context.Context, q *GetNote) (*Note, error)
}

type GetNoteHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	readModel            GetNoteReadModel
}

func NewGetNoteHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	readModel GetNoteReadModel,
) *GetNoteHandler {
	return &GetNoteHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		readModel:            readModel,
	}
}

var ProvideGetNoteHandler = NewGetNoteHandler

func (h *GetNoteHandler) Handle(ctx context.Context, query *GetNote) (*Note, error) {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		workspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read note %s", query.UserID, query.ID),
		)
	}
	return h.readModel.GetNote(ctx, query)
}
