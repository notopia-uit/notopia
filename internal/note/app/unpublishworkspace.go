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
	// WARN: Handler is incomplete - domain.Workspace has no Unpublish() method.
	// TODO: domain.Workspace has no Unpublish() method. Add a published field and
	// Unpublish() method to domain.Workspace, then call workspace.Unpublish() here before Save.
	// This mirrors PublishWorkspace handler - requires same domain.Workspace.published field addition.
	// Steps:
	// 1. Add `published bool` field to domain.Workspace struct (done with Publish handler)
	// 2. Add Unpublish() method: func (w *Workspace) Unpublish() { w.published = false }
	// 3. Update Workspace.Unmarshal() (done with Publish handler)
	// 4. Update persistence layer (done with Publish handler)
	// 5. Implement this handler to call workspace.Unpublish(), add event, and save
	_, err := h.workspacerepo.GetBySlug(ctx, cmd.Slug, false)
	if err != nil {
		return domain.NewErrWorkspaceNotFound(cmd.Slug, err)
	}
	// TODO: workspace.Unpublish() not yet implemented
	return nil
}
