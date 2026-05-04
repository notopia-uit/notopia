package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetMyWorkspaces struct {
	UserID string
}

type GetMyWorkspacesHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        GetWorkspacesReadModel
}

func NewGetMyWorkspacesHandler(
	authorizationSvc AuthorizationSvc,
	readModel GetWorkspacesReadModel,
) *GetMyWorkspacesHandler {
	return &GetMyWorkspacesHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideGetMyWorkspacesHandler = NewGetMyWorkspacesHandler

func (h *GetMyWorkspacesHandler) Handle(ctx context.Context, query *GetMyWorkspaces) ([]UserWorkspace, error) {
	slog.DebugContext(ctx, "Handling get my workspaces query", slog.String("user_id", query.UserID))
	authorizationUserWorkspaces, err := h.authorizationSvc.GetUserWorkspaces(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	workspaceIDs := make([]uuid.UUID, len(authorizationUserWorkspaces))
	for i, uw := range authorizationUserWorkspaces {
		workspaceIDs[i] = uw.ID
	}
	workspaces, err := h.readModel.GetWorkspaces(ctx, workspaceIDs)
	if err != nil {
		return nil, err
	}
	if len(workspaces) != len(authorizationUserWorkspaces) {
		return nil, errs.NewInternal(fmt.Sprintf("number of workspaces found (%d) does not match number of user workspaces (%d) for user %s", len(workspaces), len(authorizationUserWorkspaces), query.UserID))
	}
	workspaceIDToIndex := make(map[uuid.UUID]int, len(workspaces))
	for i := range workspaces {
		workspaceIDToIndex[workspaces[i].ID] = i
	}
	userWorkspaces := make([]UserWorkspace, len(authorizationUserWorkspaces))
	for i, auw := range authorizationUserWorkspaces {
		wsIndex, ok := workspaceIDToIndex[auw.ID]
		if !ok {
			return nil, errs.NewInternal(fmt.Sprintf("workspace with ID %s not found for user %s", auw.ID, query.UserID))
		}
		userWorkspaces[i] = UserWorkspace{
			Workspace: workspaces[wsIndex],
			Role:      auw.Role,
		}
	}
	slog.InfoContext(ctx, "Get my workspaces query completed", slog.Int("count", len(userWorkspaces)))
	return userWorkspaces, nil
}
