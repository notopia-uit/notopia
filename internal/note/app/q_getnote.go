package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

// FIXME: query don't use domain repo, because it is used for triggering business logic
type GetNote struct {
	ID             uuid.UUID
	ExcludeTrashed bool

	UserID string
}

type GetNoteHandler struct {
	authorizationSvc AuthorizationSvc
	noteRepo         domain.NoteRepo
	readModel        GetNoteReadModel
}

func NewGetNoteHandler(
	authorizationSvc AuthorizationSvc,
	noteRepo domain.NoteRepo,
	readModel GetNoteReadModel,
) *GetNoteHandler {
	return &GetNoteHandler{
		authorizationSvc: authorizationSvc,
		noteRepo:         noteRepo,
		readModel:        readModel,
	}
}

var ProvideGetNoteHandler = NewGetNoteHandler

type GetNoteQuery commonhandler.Query[GetNote, *Note]

var _ GetNoteQuery = (*GetNoteHandler)(nil)

func (h *GetNoteHandler) Handle(ctx context.Context, query *GetNote) (*Note, error) {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
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
	note, err := h.readModel.GetNote(ctx, &GetNoteReadModelParams{
		ID:             query.ID,
		ExcludeTrashed: query.ExcludeTrashed,
	})
	return note, err
}
