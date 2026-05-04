package app

import (
	"context"
	"log/slog"
)

type CheckWorkspaceSlugExists struct {
	Slug string
}

type CheckWorkspaceSlugExistsHandler struct {
	readModel CheckWorkspaceSlugExistsReadModel
}

func NewCheckWorkspaceSlugExistsHandler(readModel CheckWorkspaceSlugExistsReadModel) *CheckWorkspaceSlugExistsHandler {
	return &CheckWorkspaceSlugExistsHandler{readModel: readModel}
}

var ProvideCheckWorkspaceSlugExistsHandler = NewCheckWorkspaceSlugExistsHandler

func (h *CheckWorkspaceSlugExistsHandler) Handle(ctx context.Context, query *CheckWorkspaceSlugExists) (bool, error) {
	slog.DebugContext(ctx, "checking workspace slug exists", slog.String("slug", query.Slug))
	exists, err := h.readModel.CheckWorkspaceSlugExists(ctx, query.Slug)
	if err == nil {
		slog.DebugContext(
			ctx, "workspace slug check completed",
			slog.String("slug", query.Slug),
			slog.Bool("exists", exists),
		)
	}
	return exists, err
}
