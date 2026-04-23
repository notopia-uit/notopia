package app

import (
	"context"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type PublishWorkspace struct {
	Slug   string
	UserID string
}

type PublishWorkspaceHandler struct {
	workspaceRepo domain.WorkspaceRepo
}

func NewPublishWorkspaceHandler(workspaceRepo domain.WorkspaceRepo) *PublishWorkspaceHandler {
	return &PublishWorkspaceHandler{workspaceRepo: workspaceRepo}
}

var ProvidePublishWorkspaceHandler = NewPublishWorkspaceHandler

func (h *PublishWorkspaceHandler) Handle(ctx context.Context, cmd *PublishWorkspace) error {
	slog.DebugContext(ctx, "publishing workspace", slog.String("slug", cmd.Slug), slog.String("user_id", cmd.UserID))
	return nil
}
