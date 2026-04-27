package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceSearchToken struct {
	WorkspaceID uuid.UUID
	UserID      string
}

type GetWorkspaceSearchTokenHandler struct {
	authorizationSvc AuthorizationSvc
	searchSvc        SearchSvc
}

func NewGetWorkspaceSearchTokenHandler(
	authorizationSvc AuthorizationSvc,
	searchSvc SearchSvc,
) *GetWorkspaceSearchTokenHandler {
	return &GetWorkspaceSearchTokenHandler{
		authorizationSvc: authorizationSvc,
		searchSvc:        searchSvc,
	}
}

var ProvideGetWorkspaceSearchTokenHandler = NewGetWorkspaceSearchTokenHandler

func (h *GetWorkspaceSearchTokenHandler) Handle(ctx context.Context, cmd *GetWorkspaceSearchToken) (SearchToken, error) {
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(ctx, cmd.UserID, cmd.WorkspaceID, WorkspacePermissionRead)
	if err != nil {
		return SearchToken{}, err
	}

	if !hasPermission {
		return SearchToken{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to search in workspace %s", cmd.UserID, cmd.WorkspaceID),
		)
	}

	return h.searchSvc.GenerateWorkspaceToken(ctx, cmd.WorkspaceID)
}
