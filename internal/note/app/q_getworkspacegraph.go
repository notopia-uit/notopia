package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceGraph struct {
	ID            uuid.UUID
	IgnoreOrphans bool

	UserID string
}

type GetWorkspaceGraphReadModel interface {
	GetWorkspaceGraph(ctx context.Context, q *GetWorkspaceGraph) (*Graph, error)
}

type GetWorkspaceGraphHandler struct {
	authorizationService AuthorizationService
	readModel            GetWorkspaceGraphReadModel
}

func NewGetWorkspaceGraphHandler(
	authorizationService AuthorizationService,
	readModel GetWorkspaceGraphReadModel,
) *GetWorkspaceGraphHandler {
	return &GetWorkspaceGraphHandler{
		authorizationService: authorizationService,
		readModel:            readModel,
	}
}

var ProvideGetWorkspaceGraphHandler = NewGetWorkspaceGraphHandler

func (h *GetWorkspaceGraphHandler) Handle(ctx context.Context, query *GetWorkspaceGraph) (*Graph, error) {
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		query.ID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace graph %s", query.UserID, query.ID),
		)
	}
	return h.readModel.GetWorkspaceGraph(ctx, query)
}
