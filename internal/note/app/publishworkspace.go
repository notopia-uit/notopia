package app

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
	// WARN: Handler is incomplete - domain.Workspace has no Publish() method.
	// TODO: domain.Workspace has no Publish() method. Add a published field and
	// Publish() method to domain.Workspace, then call workspace.Publish() here before Save.
	// Steps:
	// 1. Add `published bool` field to domain.Workspace struct
	// 2. Add Publish() method: func (w *Workspace) Publish() { w.published = true }
	// 3. Update Workspace.Unmarshal() to accept published parameter
	// 4. Update persistence layer to store/retrieve published field
	// 5. Implement this handler to call workspace.Publish(), add event, and save
	// TODO: workspace.Publish() not yet implemented
	return nil
}
