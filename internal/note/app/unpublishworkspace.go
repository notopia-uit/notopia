package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type UnpublishWorkspace struct {
	Slug string
}

type UnpublishWorkspaceHandler struct {
	workspacerepo domain.WorkspaceRepo
}

func NewUnpublishWorkspaceHandler(workspacerepo domain.WorkspaceRepo) *UnpublishWorkspaceHandler {
	return &UnpublishWorkspaceHandler{workspacerepo: workspacerepo}
}

var ProvideUnpublishWorkspaceHandler = NewUnpublishWorkspaceHandler

func (h *UnpublishWorkspaceHandler) Handle(ctx context.Context, cmd *UnpublishWorkspace) error {
	// TODO: domain.Workspace has no Unpublish() method. Add a published field and
	// Unpublish() method to domain.Workspace, then call workspace.Unpublish() here before Save.
	_, err := h.workspacerepo.GetBySlug(ctx, cmd.Slug, false)
	if err != nil {
		return domain.NewErrWorkspaceNotFound(cmd.Slug, err)
	}
	return nil
}
