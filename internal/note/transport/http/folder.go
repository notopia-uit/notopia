package http

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) CreateFolder(
	ctx context.Context,
	request note.CreateFolderRequestObject,
) (note.CreateFolderResponseObject, error) {
	return &note.CreateFolder201Response{}, nil
}

func (h *StrictHandler) RenameFolder(
	ctx context.Context,
	request note.RenameFolderRequestObject,
) (note.RenameFolderResponseObject, error) {
	return &note.RenameFolder204Response{}, nil
}
