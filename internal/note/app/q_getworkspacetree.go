package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceTree struct {
	WorkspaceID    uuid.UUID
	RootFolderID   uuid.UUID
	IncludeTrashed bool
	Depth          uint

	UserID string
}

type GetWorkspaceTreeHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        GetWorkspaceTreeReadModel
}

func NewGetWorkspaceTreeHandler(
	authorizationSvc AuthorizationSvc,
	readModel GetWorkspaceTreeReadModel,
) *GetWorkspaceTreeHandler {
	return &GetWorkspaceTreeHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideGetWorkspaceTreeHandler = NewGetWorkspaceTreeHandler

func (h *GetWorkspaceTreeHandler) Handle(ctx context.Context, query *GetWorkspaceTree) (WorkspaceTreeFolder, error) {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		query.WorkspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return WorkspaceTreeFolder{}, err
	}
	if !hasPermission {
		return WorkspaceTreeFolder{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace tree %s", query.UserID, query.WorkspaceID),
		)
	}
	tree, err := h.readModel.GetWorkspaceTree(ctx, &GetWorkspaceTreeReadModelParams{
		WorkspaceID:    query.WorkspaceID,
		RootFolderID:   query.RootFolderID,
		IncludeTrashed: query.IncludeTrashed,
		Depth:          query.Depth,
	})
	if err != nil {
		return WorkspaceTreeFolder{}, err
	}
	return tree, nil
}
