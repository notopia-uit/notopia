package app

import (
	"context"

	"github.com/google/uuid"
)

type ShowTrash struct {
	WorkspaceID uuid.UUID
}

type ShowTrashReadModel interface {
	ShowTrash(ctx context.Context, q *ShowTrash) (*Trash, error)
}

type ShowTrashHandler struct {
	readModel ShowTrashReadModel
}

func NewShowTrashHandler(readModel ShowTrashReadModel) *ShowTrashHandler {
	return &ShowTrashHandler{readModel: readModel}
}

var ProvideShowTrashHandler = NewShowTrashHandler

func (h *ShowTrashHandler) Handle(ctx context.Context, query *ShowTrash) (*Trash, error) {
	return h.readModel.ShowTrash(ctx, query)
}
