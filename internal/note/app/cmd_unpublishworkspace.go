package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
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

type UnpublishWorkspaceCmd commonhandler.Cmd[UnpublishWorkspace]

var _ UnpublishWorkspaceCmd = (*UnpublishWorkspaceHandler)(nil)

func (h *UnpublishWorkspaceHandler) Handle(ctx context.Context, cmd *UnpublishWorkspace) error {
	return nil
}
