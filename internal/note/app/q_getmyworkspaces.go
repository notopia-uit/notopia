package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetMyWorkspaces struct {
	UserID string
}

type GetMyWorkspacesReadModel interface {
	GetWorkspacesByIDs(ctx context.Context, ids []uuid.UUID) ([]Workspace, error)
}

type GetMyWorkspacesHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        GetMyWorkspacesReadModel
}

func NewGetMyWorkspacesHandler(
	authorizationSvc AuthorizationSvc,
	readModel GetMyWorkspacesReadModel,
) *GetMyWorkspacesHandler {
	return &GetMyWorkspacesHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideGetMyWorkspacesHandler = NewGetMyWorkspacesHandler

func (h *GetMyWorkspacesHandler) Handle(ctx context.Context, query *GetMyWorkspaces) ([]UserWorkspace, error) {
	authorizationUserWorkspaces, err := h.authorizationSvc.GetUserWorkspaces(ctx, query.UserID)
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
	workspaceIDToIndex := make(map[uuid.UUID]int, len(workspaces))
	for i := range workspaces {
		workspaceIDToIndex[workspaces[i].ID] = i
	}
	userWorkspaces := make([]UserWorkspace, 0, len(authorizationUserWorkspaces))
	for _, auw := range authorizationUserWorkspaces {
		wsIndex, ok := workspaceIDToIndex[auw.ID]
		if !ok {
			return nil, errs.NewInternal("workspace not found for user workspace")
		}
		userWorkspaces = append(userWorkspaces, UserWorkspace{
			Workspace: &workspaces[wsIndex],
			Role:      auw.Role,
		})
	}
	return userWorkspaces, nil
}
