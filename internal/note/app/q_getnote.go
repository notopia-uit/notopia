package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
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

func (h *GetNoteHandler) Handle(ctx context.Context, query *GetNote) (*Note, error) {
	slog.DebugContext(
		ctx, "getting note",
		slog.String("note_id", query.ID.String()),
		slog.Bool("exclude_trashed", query.ExcludeTrashed),
		slog.String("user_id", query.UserID),
	)
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	slog.DebugContext(
		ctx, "checking permission",
		slog.String("user_id", query.UserID),
		slog.String("workspace_id", workspaceID.String()),
		slog.String("permission", "read"),
	)
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
	slog.DebugContext(
		ctx, "permission granted",
		slog.String("user_id", query.UserID),
		slog.String("note_id", query.ID.String()),
	)
	note, err := h.readModel.GetNote(ctx, &GetNoteReadModelParams{
		ID:             query.ID,
		ExcludeTrashed: query.ExcludeTrashed,
	})
	if err == nil && note != nil {
		slog.InfoContext(ctx, "note retrieved successfully", slog.String("note_id", query.ID.String()))
	}
	return note, err
}
