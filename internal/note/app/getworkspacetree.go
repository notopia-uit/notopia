package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceTree struct {
	WorkspaceID    uuid.UUID
	RootFolderID   *uuid.UUID
	IncludeTrashed bool
	Depth          *uint
}

type GetWorkspaceTreeReadModel interface {
	GetWorkspaceTree(ctx context.Context, q *GetWorkspaceTree) (*WorkspaceTreeFolder, errs.Error)
}

type GetWorkspaceTreeHandler struct {
	readModel GetWorkspaceTreeReadModel
}

func NewGetWorkspaceTreeHandler(readModel GetWorkspaceTreeReadModel) *GetWorkspaceTreeHandler {
	return &GetWorkspaceTreeHandler{readModel: readModel}
}

var ProvideGetWorkspaceTreeHandler = NewGetWorkspaceTreeHandler

func (h *GetWorkspaceTreeHandler) Handle(ctx context.Context, query *GetWorkspaceTree) (*WorkspaceTreeFolder, errs.Error) {
	// TODO: Authorize
	return h.readModel.GetWorkspaceTree(ctx, query)
}
