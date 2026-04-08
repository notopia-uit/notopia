package app

import (
	"context"
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceBySlug struct {
	Slug string

	UserID string
}

type WorkspaceBySlugReadModel interface {
	GetWorkspaceBySlug(ctx context.Context, q *GetWorkspaceBySlug) (*Workspace, error)
}

type GetWorkspaceHandler struct {
	authorizationService AuthorizationService
	readModel            WorkspaceBySlugReadModel
}

func NewGetWorkspaceBySlugHandler(
	authorizationService AuthorizationService,
	readModel WorkspaceBySlugReadModel,
) *GetWorkspaceHandler {
	return &GetWorkspaceHandler{
		authorizationService: authorizationService,
		readModel:            readModel,
	}
}

var ProvideGetWorkspaceBySlugHandler = NewGetWorkspaceBySlugHandler

func (h *GetWorkspaceHandler) Handle(ctx context.Context, query *GetWorkspaceBySlug) (*Workspace, error) {
	workspace, err := h.readModel.GetWorkspaceBySlug(ctx, query)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, nil
	}
	hasPermission, err := h.authorizationService.HasWorkspacePermission(
		ctx,
		query.UserID,
		workspace.ID,
		WorkspacePermissionRead,
	)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace %s", query.UserID, workspace.ID),
		)
	}
	return workspace, nil
}
