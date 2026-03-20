package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type RenameWorkspace struct {
	ID     uuid.UUID
	UserID string
	Name   string
}

type RenameWorkspaceHandler struct {
	authorizationService AuthorizationService
	workspacerepo        domain.WorkspaceRepo
}

func NewRenameWorkspaceHandler(
	authorizationService AuthorizationService,
	workspacerepo domain.WorkspaceRepo,
) *RenameWorkspaceHandler {
	return &RenameWorkspaceHandler{
		authorizationService: authorizationService,
		workspacerepo:        workspacerepo,
	}
}

var ProvideRenameWorkspaceHandler = NewRenameWorkspaceHandler

func (h *RenameWorkspaceHandler) Handle(ctx context.Context, cmd *RenameWorkspace) error {
	hasPermission, err := h.authorizationService.HasWorkspacePermission(
		ctx,
		cmd.UserID,
		cmd.ID,
		WorkspacePermissionEdit,
	)
	if err != nil {
		return err
	}

	if !hasPermission {
		return newErrRenameWorkspaceForbidden(cmd.UserID, cmd.ID)
	}

	workspace, err := h.workspacerepo.GetByID(ctx, cmd.ID, true)
	if err != nil {
		return err
	}
	workspace.Rename(cmd.Name)
	return h.workspacerepo.Save(ctx, workspace)
}

var ErrCodeRenameWorkspaceForbidden = "RenameWorkspace_1"

func newErrRenameWorkspaceForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to rename workspace %q", userID, workspaceID.String()),
		ErrCodeRenameWorkspaceForbidden,
		nil,
	)
}
