package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type CreateWorkspaceHandler struct {
	Enforcer *casbin.TransactionalEnforcer
}

func (h *CreateWorkspaceHandler) Handle(
	userID string,
	workspaceID uuid.UUID,
) error {
	ok, err := h.Enforcer.AddGroupingPolicy(
		formatUser(userID),
		"owner",
		formatWorkspace(workspaceID),
	)
	if err != nil {
		return newErrCreateWorkspaceFailed(userID, workspaceID)
	}
	if !ok {
		return newErrCreateWorkspaceExists(userID, workspaceID)
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
