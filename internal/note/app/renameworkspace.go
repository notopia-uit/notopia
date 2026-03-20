package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type RenameWorkspace struct {
	Slug string
	Name string
}

type RenameWorkspaceHandler struct {
	workspacerepo domain.WorkspaceRepo
}

func NewRenameWorkspaceHandler(workspacerepo domain.WorkspaceRepo) *RenameWorkspaceHandler {
	return &RenameWorkspaceHandler{workspacerepo: workspacerepo}
}

var ProvideRenameWorkspaceHandler = NewRenameWorkspaceHandler

func (h *RenameWorkspaceHandler) Handle(ctx context.Context, cmd *RenameWorkspace) error {
	workspace, err := h.workspacerepo.GetBySlug(ctx, cmd.Slug, true)
	if err != nil {
		return domain.NewErrWorkspaceNotFound(cmd.Slug, err)
	}
	workspace.Rename(cmd.Name)
	return h.workspacerepo.Save(ctx, workspace)
}
