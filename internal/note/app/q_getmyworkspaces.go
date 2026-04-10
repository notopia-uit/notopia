package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
)

type GetMyWorkspaces struct {
	UserID string
}

type GetMyWorkspacesReadModel interface {
	GetWorkspacesByIDs(ctx context.Context, ids []uuid.UUID) ([]*Workspace, error)
}

type GetMyWorkspacesHandler struct {
	authorizationService AuthorizationService
	readModel            GetMyWorkspacesReadModel
}

func NewGetMyWorkspacesHandler(
	authorizationService AuthorizationService,
	readModel GetMyWorkspacesReadModel,
) *GetMyWorkspacesHandler {
	return &GetMyWorkspacesHandler{
		authorizationService: authorizationService,
		readModel:            readModel,
	}
}

var ProvideGetMyWorkspacesHandler = NewGetMyWorkspacesHandler

func (h *GetMyWorkspacesHandler) Handle(ctx context.Context, query *GetMyWorkspaces) ([]*UserWorkspace, error) {
	authorizationUserWorkspaces, err := h.authorizationService.GetUserWorkspaces(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	workspaceIDs := make([]uuid.UUID, len(authorizationUserWorkspaces))
	for i, uw := range authorizationUserWorkspaces {
		workspaceIDs[i] = uw.ID
	}
	workspaces, err := h.readModel.GetWorkspacesByIDs(ctx, workspaceIDs)
	if err != nil {
		return nil, err
	}
	workspaceIDWorkspaceMap := make(map[uuid.UUID]*Workspace)
	for _, w := range workspaces {
		workspaceIDWorkspaceMap[w.ID] = w
	}
	userWorkspaces := make([]*UserWorkspace, len(workspaces))
	for i, auw := range authorizationUserWorkspaces {
		w, ok := workspaceIDWorkspaceMap[auw.ID]
		if !ok {
			return nil, errs.NewInternal("workspace not found for user workspace")
		}
		userWorkspaces[i] = &UserWorkspace{
			Workspace: w,
			Role:      auw.Role,
		}
	}
	return userWorkspaces, nil
}
