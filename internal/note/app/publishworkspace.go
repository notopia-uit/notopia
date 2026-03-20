package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type PublishWorkspace struct {
	Slug string
}

type PublishWorkspaceHandler struct {
	workspacerepo domain.WorkspaceRepo
}

func NewPublishWorkspaceHandler(workspacerepo domain.WorkspaceRepo) *PublishWorkspaceHandler {
	return &PublishWorkspaceHandler{workspacerepo: workspacerepo}
}

var ProvidePublishWorkspaceHandler = NewPublishWorkspaceHandler

func (h *PublishWorkspaceHandler) Handle(ctx context.Context, cmd *PublishWorkspace) error {
	// TODO: domain.Workspace has no Publish() method. Add a published field and
	// Publish() method to domain.Workspace, then call workspace.Publish() here before Save.
	_, err := h.workspacerepo.GetBySlug(ctx, cmd.Slug, true)
	if err != nil {
		return domain.NewErrWorkspaceNotFound(cmd.Slug, err)
	}
	return nil
}
