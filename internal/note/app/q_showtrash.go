package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ShowTrash struct {
	WorkspaceID uuid.UUID

	UserID string
}

type ShowTrashReadModel interface {
	ShowTrash(ctx context.Context, q *ShowTrash) (Trash, error)
}

type ShowTrashHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        ShowTrashReadModel
}

func NewShowTrashHandler(
	authorizationSvc AuthorizationSvc,
	readModel ShowTrashReadModel,
) *ShowTrashHandler {
	return &ShowTrashHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideShowTrashHandler = NewShowTrashHandler

type ShowTrashQuery commonhandler.Query[ShowTrash, Trash]

var _ ShowTrashQuery = (*ShowTrashHandler)(nil)

func (h *ShowTrashHandler) Handle(ctx context.Context, query *ShowTrash) (Trash, error) {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		query.WorkspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return Trash{}, err
	}
	if !hasPermission {
		return Trash{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read trash in workspace %s", query.UserID, query.WorkspaceID),
		)
	}
	trash, err := h.readModel.ShowTrash(ctx, query)
	if err != nil {
		return Trash{}, err
	}
	return trash, nil
}
