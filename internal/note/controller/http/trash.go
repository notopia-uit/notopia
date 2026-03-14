package http

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) RestoreTrashedWorkspaceItems(
	ctx context.Context,
	request note.RestoreTrashedWorkspaceItemsRequestObject,
) (note.RestoreTrashedWorkspaceItemsResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) ShowTrash(
	ctx context.Context,
	request note.ShowTrashRequestObject,
) (note.ShowTrashResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) TrashWorkspaceItems(
	ctx context.Context,
	request note.TrashWorkspaceItemsRequestObject,
) (note.TrashWorkspaceItemsResponseObject, error) {
	return nil, errors.New("not implemented")
}
