package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type DeleteWorkspace struct {
	ID uuid.UUID
}

type DeleteWorkspaceHandler struct {
	workspaceRepo domain.WorkspaceRepo
}

func NewDeleteWorkspaceHandler(workspaceRepo domain.WorkspaceRepo) *DeleteWorkspaceHandler {
	return &DeleteWorkspaceHandler{workspaceRepo: workspaceRepo}
}

var ProvideDeleteWorkspaceHandler = NewDeleteWorkspaceHandler

func (h *DeleteWorkspaceHandler) Handle(ctx context.Context, cmd *DeleteWorkspace) error {
	workspace, err := h.workspaceRepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	workspace.Delete()
	return h.workspaceRepo.Save(ctx, workspace)
}
