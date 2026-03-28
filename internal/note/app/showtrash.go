package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type ShowTrash struct {
	WorkspaceID uuid.UUID
}

type ShowTrashReadModel interface {
	ShowTrash(ctx context.Context, q *ShowTrash) (*Trash, errs.Error)
}

type ShowTrashHandler struct {
	readModel ShowTrashReadModel
}

func NewShowTrashHandler(readModel ShowTrashReadModel) *ShowTrashHandler {
	return &ShowTrashHandler{readModel: readModel}
}

var ProvideShowTrashHandler = NewShowTrashHandler

func (h *ShowTrashHandler) Handle(ctx context.Context, query *ShowTrash) (*Trash, errs.Error) {
	return h.readModel.ShowTrash(ctx, query)
}
