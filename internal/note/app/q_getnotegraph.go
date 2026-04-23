package app

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetNoteGraph struct {
	ID    uuid.UUID
	Depth int

	UserID string
}

type GetNoteGraphHandler struct {
	authorizationSvc AuthorizationSvc
	noteRepo         domain.NoteRepo
	readModel        GetNoteGraphReadModel
}

func NewGetNoteGraphHandler(
	authorizationSvc AuthorizationSvc,
	noteRepo domain.NoteRepo,
	readModel GetNoteGraphReadModel,
) *GetNoteGraphHandler {
	return &GetNoteGraphHandler{
		authorizationSvc: authorizationSvc,
		noteRepo:         noteRepo,
		readModel:        readModel,
	}
}

var ProvideGetNoteGraphHandler = NewGetNoteGraphHandler

func (h *GetNoteGraphHandler) Handle(ctx context.Context, query *GetNoteGraph) (Graph, error) {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		return Graph{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		workspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return Graph{}, err
	}
	if !hasPermission {
		return Graph{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read note graph %s", query.UserID, query.ID),
		)
	}
	if query.Depth <= 0 {
		query.Depth = math.MaxInt
	}
	return h.readModel.GetNoteGraph(ctx, &GetNoteGraphReadModelParams{
		ID:    query.ID,
		Depth: query.Depth,
	})
}
