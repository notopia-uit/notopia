package http

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) CreateFolder(
	ctx context.Context,
	request note.CreateFolderRequestObject,
) (note.CreateFolderResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) RenameFolder(
	ctx context.Context,
	request note.RenameFolderRequestObject,
) (note.RenameFolderResponseObject, error) {
	return nil, errors.New("not implemented")
}
