package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetNoteLinks struct {
	ID            uuid.UUID
	OutgoingLinks bool
	Backlinks     bool

	UserID string
}

type GetNoteLinksHandler struct {
	authorizationSvc AuthorizationSvc
	noteRepo         domain.NoteRepo
	readModel        GetNoteLinksReadModel
}

func NewGetNoteLinksHandler(
	authorizationSvc AuthorizationSvc,
	noteRepo domain.NoteRepo,
	readModel GetNoteLinksReadModel,
) *GetNoteLinksHandler {
	return &GetNoteLinksHandler{
		authorizationSvc: authorizationSvc,
		noteRepo:         noteRepo,
		readModel:        readModel,
	}
}

var ProvideGetNoteLinksHandler = NewGetNoteLinksHandler

func (h *GetNoteLinksHandler) Handle(ctx context.Context, query *GetNoteLinks) (NoteLinkResult, error) {
	slog.DebugContext(ctx, "Handling get note links query", slog.String("note_id", query.ID.String()))
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get workspace ID for note", slog.String("note_id", query.ID.String()), slog.Any("error", err))
		return NoteLinkResult{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		workspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check permission", slog.String("user_id", query.UserID), slog.String("workspace_id", workspaceID.String()), slog.Any("error", err))
		return NoteLinkResult{}, err
	}
	if !hasPermission {
		slog.WarnContext(ctx, "permission denied", slog.String("user_id", query.UserID), slog.String("note_id", query.ID.String()))
		return NoteLinkResult{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read note links %s", query.UserID, query.ID),
		)
	}
	result, err := h.readModel.GetNoteLinks(ctx, &GetNoteLinksReadModelParams{
		ID:            query.ID,
		OutgoingLinks: query.OutgoingLinks,
		Backlinks:     query.Backlinks,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to get note links", slog.String("note_id", query.ID.String()), slog.Any("error", err))
		return NoteLinkResult{}, err
	}
	slog.InfoContext(ctx, "Get note links query completed", slog.String("note_id", query.ID.String()))
	return result, nil
}
