package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type GetMyWorkspaces struct {
	UserID string
}

// TODO: will inject authentik client to get, talked via IdentityService
type GetMyWorkspacesHandler struct {
	authorizationService AuthorizationService
}

func NewGetMyWorkspacesHandler(authorizationService AuthorizationService) *GetMyWorkspacesHandler {
	return &GetMyWorkspacesHandler{authorizationService: authorizationService}
}

var ProvideGetMyWorkspacesHandler = NewGetMyWorkspacesHandler

func (h *GetMyWorkspacesHandler) Handle(ctx context.Context, query *GetMyWorkspaces) ([]*UserWorkspace, error) {
	// authorizationUserWorkspaces
	_, err := h.authorizationService.GetUserWorkspaces(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	return nil, errs.Unimplemented
}
