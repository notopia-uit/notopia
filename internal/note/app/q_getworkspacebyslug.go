package app

import (
	"context"
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetWorkspaceBySlug struct {
	Slug string

	UserID string
}

type GetWorkspaceBySlugHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        WorkspaceBySlugReadModel
}

func NewGetWorkspaceBySlugHandler(
	authorizationSvc AuthorizationSvc,
	readModel WorkspaceBySlugReadModel,
) *GetWorkspaceBySlugHandler {
	return &GetWorkspaceBySlugHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideGetWorkspaceBySlugHandler = NewGetWorkspaceBySlugHandler

type GetWorkspaceBySlugQuery commonhandler.Query[GetWorkspaceBySlug, Workspace]

var _ GetWorkspaceBySlugQuery = (*GetWorkspaceBySlugHandler)(nil)

func (h *GetWorkspaceBySlugHandler) Handle(ctx context.Context, query *GetWorkspaceBySlug) (Workspace, error) {
	workspace, err := h.readModel.Handle(ctx, query.Slug)
	if err != nil {
		return Workspace{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(
		ctx,
		query.UserID,
		workspace.ID,
		WorkspacePermissionRead,
	)
	if err != nil {
		return Workspace{}, err
	}
	if !hasPermission {
		return Workspace{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace %s", query.UserID, workspace.ID),
		)
	}
	return workspace, nil
}
