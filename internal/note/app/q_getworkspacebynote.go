package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceByNote struct {
	NoteID uuid.UUID
	UserID string
}

type GetWorkspaceByNoteHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        GetWorkspaceByNoteReadModel
}

func NewGetWorkspaceByNoteHandler(
	authorizationSvc AuthorizationSvc,
	readModel GetWorkspaceByNoteReadModel,
) *GetWorkspaceByNoteHandler {
	return &GetWorkspaceByNoteHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideGetWorkspaceByNoteHandler = NewGetWorkspaceByNoteHandler

func (h *GetWorkspaceByNoteHandler) Handle(ctx context.Context, query *GetWorkspaceByNote) (Workspace, error) {
	slog.DebugContext(ctx, "Handling get workspace by note query", slog.String("note_id", query.NoteID.String()))
	workspace, err := h.readModel.GetWorkspaceByNoteID(ctx, query.NoteID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get workspace by note ID", slog.String("note_id", query.NoteID.String()), slog.Any("error", err))
		return Workspace{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(ctx, query.UserID, workspace.ID, WorkspacePermissionRead)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check permission", slog.String("user_id", query.UserID), slog.String("workspace_id", workspace.ID.String()), slog.Any("error", err))
		return Workspace{}, err
	}
	if !hasPermission {
		slog.WarnContext(ctx, "permission denied", slog.String("user_id", query.UserID), slog.String("workspace_id", workspace.ID.String()))
		return Workspace{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace %s", query.UserID, workspace.ID),
		)
	}
	slog.InfoContext(ctx, "Get workspace by note query completed", slog.String("workspace_id", workspace.ID.String()))
	return workspace, nil
}
