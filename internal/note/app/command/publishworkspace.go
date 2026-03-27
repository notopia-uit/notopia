package command

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type PublishWorkspace struct {
	Slug string
}

type PublishWorkspaceHandler struct {
	workspaceRepo domain.WorkspaceRepo
}

func NewPublishWorkspaceHandler(workspaceRepo domain.WorkspaceRepo) *PublishWorkspaceHandler {
	return &PublishWorkspaceHandler{workspaceRepo: workspaceRepo}
}

var ProvidePublishWorkspaceHandler = NewPublishWorkspaceHandler

func (h *PublishWorkspaceHandler) Handle(ctx context.Context, cmd *PublishWorkspace) error {
	return nil
}
