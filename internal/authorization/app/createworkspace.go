package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type CreateWorkspace struct {
	UserID      string
	WorkspaceID uuid.UUID
}

type CreateWorkspaceHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewCreateWorkspaceHandler(enforcer *casbin.TransactionalEnforcer) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{enforcer: enforcer}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

func (h *CreateWorkspaceHandler) Handle(params CreateWorkspace) error {
	ok, err := h.enforcer.AddGroupingPolicy(
		formatUser(params.UserID),
		"owner",
		formatWorkspace(params.WorkspaceID),
	)
	if err != nil {
		return newErrCreateWorkspaceFailed(params.UserID, params.WorkspaceID)
	}
	if !ok {
		return newErrCreateWorkspaceExists(params.UserID, params.WorkspaceID)
	}
	return nil
}

var (
	ErrCodeCreateWorkspaceFailed = "CreateWorkspace_1"
	ErrCodeCreateWorkspaceExists = "CreateWorkspace_2"
)

func newErrCreateWorkspaceFailed(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewInternal(
		fmt.Sprintf("Failed to create workspace %q for user %q", workspaceID.String(), userID),
		ErrCodeCreateWorkspaceFailed,
		nil,
	)
}

func newErrCreateWorkspaceExists(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewConflict(
		fmt.Sprintf("Workspace %q already exists for user %q", workspaceID.String(), userID),
		ErrCodeCreateWorkspaceExists,
		nil,
	)
}
