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

type GetWorkspaceTreeReadModel interface {
	GetWorkspaceTree(ctx context.Context, q *GetWorkspaceTree) (*WorkspaceTreeFolder, error)
}

type GetWorkspaceTreeHandler struct {
	authorizationService AuthorizationService
	readModel            GetWorkspaceTreeReadModel
}

func NewGetWorkspaceTreeHandler(
	authorizationService AuthorizationService,
	readModel GetWorkspaceTreeReadModel,
) *GetWorkspaceTreeHandler {
	return &GetWorkspaceTreeHandler{
		authorizationService: authorizationService,
		readModel:            readModel,
	}
}

var ProvideGetWorkspaceTreeHandler = NewGetWorkspaceTreeHandler

func (h *GetWorkspaceTreeHandler) Handle(ctx context.Context, query *GetWorkspaceTree) (*WorkspaceTreeFolder, error) {
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		query.WorkspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace tree %s", query.UserID, query.WorkspaceID),
		)
	}
	return h.readModel.GetWorkspaceTree(ctx, query)
}
