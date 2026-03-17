package command

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type DeleteWorkspace struct {
	Slug string
}

type DeleteWorkspaceHandler struct {
	workspacerepo domain.WorkspaceRepo
}

func NewDeleteWorkspaceHandler(workspacerepo domain.WorkspaceRepo) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{workspacerepo: workspacerepo}
}

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd *DeleteWorkspace) error {
	workspace, err := h.workspacerepo.GetBySlug(ctx, cmd.Slug, true)
	if err != nil {
		return domain.NewErrWorkspaceNotFound(cmd.Slug, err)
	}
	workspace.Delete()
	return h.workspacerepo.Save(ctx, workspace)
}
