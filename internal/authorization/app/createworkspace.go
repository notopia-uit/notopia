package app

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CreateWorkspace struct {
	OwnerID     string
	WorkspaceID uuid.UUID
}

type CreateWorkspaceHandler struct {
	enforcer *casbin.TransactionalEnforcer
}

func NewCreateWorkspaceHandler(enforcer *casbin.TransactionalEnforcer) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{enforcer: enforcer}
}

var ProvideCreateWorkspaceHandler = NewCreateWorkspaceHandler

type CreateWorkspaceCmd commonhandler.Cmd[CreateWorkspace]

var _ CreateWorkspaceCmd = (*CreateWorkspaceHandler)(nil)

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, params *CreateWorkspace) error {
	ok, err := h.enforcer.AddGroupingPolicy(
		formatUser(params.OwnerID),
		WorkspaceRoleOwner.String(),
		formatWorkspace(params.WorkspaceID),
	)
	if err != nil {
		return errs.NewCasbinInternalError(err)
	}
	if !ok {
		return errs.NewCreateWorkspaceExists(params.OwnerID, params.WorkspaceID)
	}
	return nil
}
