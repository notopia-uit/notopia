package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceBySlug struct {
	Slug string

	UserID string
}

type GetWorkspaceHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        WorkspaceBySlugReadModel
}

func NewGetWorkspaceBySlugHandler(
	authorizationSvc AuthorizationSvc,
	readModel WorkspaceBySlugReadModel,
) *GetWorkspaceHandler {
	return &GetWorkspaceHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideGetWorkspaceBySlugHandler = NewGetWorkspaceBySlugHandler

func (h *GetWorkspaceHandler) Handle(ctx context.Context, query *GetWorkspaceBySlug) (Workspace, error) {
	slog.DebugContext(ctx, "Handling get workspace by slug query", slog.String("slug", query.Slug))
	workspace, err := h.readModel.GetWorkspaceBySlug(ctx, query.Slug)
	if err != nil {
		return Workspace{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(
		ctx,
		query.UserID,
		workspace.ID,
		WorkspacePermissionRead,
	)
	if err != nil {
		return Workspace{}, err
	}
	if !hasPermission {
		return Workspace{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace %s", query.UserID, workspace.ID),
		)
	}
	slog.InfoContext(ctx, "Get workspace by slug query completed", slog.String("workspace_id", workspace.ID.String()))
	return workspace, nil
}
