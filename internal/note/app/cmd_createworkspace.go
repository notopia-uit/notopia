package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CreateWorkspace struct {
	ID      uuid.UUID
	Name    string
	Slug    string
	OwnerID string
}

type CreateWorkspaceHandler struct {
	uow              domain.UnitOfWork
	authorizationSvc AuthorizationSvc
}

func NewCreateWorkspaceHandler(
	uow domain.UnitOfWork,
	authorizationSvc AuthorizationSvc,
) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{
		uow:              uow,
		authorizationSvc: authorizationSvc,
	}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

type CreateWorkspaceCmd commonhandler.Cmd[CreateWorkspace]

var _ CreateWorkspaceCmd = (*CreateWorkspaceHandler)(nil)

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, cmd *CreateWorkspace) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		workspaceRepo := r.Workspace()
		folderRepo := r.Folder()
		slugExisted, err := workspaceRepo.CheckSlugExists(ctx, cmd.Slug)
		if err != nil {
			return err
		}
		if slugExisted {
			return errs.NewWorkspaceSlugAlreadyExists(cmd.Slug, nil)
		}
		rootFolderID, err := uuid.NewV7()
		if err != nil {
			return errs.NewInternalGenerateID(err)
		}
		workspace, err := domain.NewWorkspace(cmd.ID, cmd.Name, cmd.Slug)
		if err != nil {
			return err
		}
		rootFolder, err := domain.NewFolder(rootFolderID, cmd.Name, "", cmd.ID, domain.FolderHierarchy{}, cmd.OwnerID)
		if err != nil {
			return err
		}
		if err := workspaceRepo.Save(ctx, workspace); err != nil {
			return err
		}
		if err := folderRepo.Save(ctx, rootFolder); err != nil {
			return err
		}
		// TODO: May use saga for better perf, no bottle neck
		// TODO: from @coderabbitai:
		//	This is a cross-service side effect inside uow.Execute.
		//	If the auth call succeeds and the DB commit later fails, note and authorization will diverge.
		//	Trigger it after commit or via an outbox/after-commit hook instead.
		return h.authorizationSvc.CreateWorkspaceWithOwner(ctx, cmd.OwnerID, workspace.ID())
	})
}
