package app

import (
	"context"
	"fmt"
	"log/slog"
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
	slog.DebugContext(ctx, "Handling get note graph query", slog.String("note_id", query.ID.String()), slog.Int("depth", query.Depth))
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get workspace ID for note", slog.String("note_id", query.ID.String()), slog.Any("error", err))
		return Graph{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		workspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check permission", slog.String("user_id", query.UserID), slog.String("workspace_id", workspaceID.String()), slog.Any("error", err))
		return Graph{}, err
	}
	if !hasPermission {
		slog.WarnContext(ctx, "permission denied", slog.String("user_id", query.UserID), slog.String("note_id", query.ID.String()))
		return Graph{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read note graph %s", query.UserID, query.ID),
		)
	}
	if query.Depth <= 0 {
		query.Depth = math.MaxInt
	}
	graph, err := h.readModel.GetNoteGraph(ctx, &GetNoteGraphReadModelParams{
		ID:    query.ID,
		Depth: query.Depth,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to get note graph", slog.String("note_id", query.ID.String()), slog.Any("error", err))
		return Graph{}, err
	}
	slog.InfoContext(ctx, "Get note graph query completed", slog.String("note_id", query.ID.String()))
	return graph, nil
}
