package app

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type UnpublishWorkspace struct {
	Slug string
}

type UnpublishWorkspaceHandler struct {
	workspaceRepo domain.WorkspaceRepo
}

func NewUnpublishWorkspaceHandler(workspaceRepo domain.WorkspaceRepo) *UnpublishWorkspaceHandler {
	return &UnpublishWorkspaceHandler{workspaceRepo: workspaceRepo}
}

var ProvideUnpublishWorkspaceHandler = NewUnpublishWorkspaceHandler

func (h *UnpublishWorkspaceHandler) Handle(ctx context.Context, cmd *UnpublishWorkspace) error {
	slog.DebugContext(ctx, "unpublishing workspace", slog.String("slug", cmd.Slug))
	return nil
}
