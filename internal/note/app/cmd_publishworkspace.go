package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
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

type PublishWorkspaceCmd commonhandler.Cmd[PublishWorkspace]

var _ PublishWorkspaceCmd = (*PublishWorkspaceHandler)(nil)

func (h *PublishWorkspaceHandler) Handle(ctx context.Context, cmd *PublishWorkspace) error {
	return nil
}
