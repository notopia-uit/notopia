package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type ShowTrash struct {
	WorkspaceID uuid.UUID

	UserID string
}

type ShowTrashReadModel interface {
	ShowTrash(ctx context.Context, q *ShowTrash) (Trash, error)
}

type ShowTrashHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        ShowTrashReadModel
}

func NewShowTrashHandler(
	authorizationSvc AuthorizationSvc,
	readModel ShowTrashReadModel,
) *ShowTrashHandler {
	return &ShowTrashHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideShowTrashHandler = NewShowTrashHandler

func (h *ShowTrashHandler) Handle(ctx context.Context, query *ShowTrash) (Trash, error) {
	slog.DebugContext(ctx, "Handling show trash query", slog.String("workspace_id", query.WorkspaceID.String()))
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		query.WorkspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check permission", slog.String("user_id", query.UserID), slog.String("workspace_id", query.WorkspaceID.String()), slog.Any("error", err))
		return Trash{}, err
	}
	if !hasPermission {
		slog.WarnContext(ctx, "permission denied", slog.String("user_id", query.UserID), slog.String("workspace_id", query.WorkspaceID.String()))
		return Trash{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read trash in workspace %s", query.UserID, query.WorkspaceID),
		)
	}
	trash, err := h.readModel.ShowTrash(ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, "failed to show trash", slog.String("workspace_id", query.WorkspaceID.String()), slog.Any("error", err))
		return Trash{}, err
	}
	slog.InfoContext(ctx, "Show trash query completed", slog.String("workspace_id", query.WorkspaceID.String()))
	return trash, nil
}
