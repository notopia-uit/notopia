package http

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) GetRevisions(
	ctx context.Context,
	request note.GetRevisionsRequestObject,
) (note.GetRevisionsResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) DeleteRevision(
	ctx context.Context,
	request note.DeleteRevisionRequestObject,
) (note.DeleteRevisionResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetRevision(
	ctx context.Context,
	request note.GetRevisionRequestObject,
) (note.GetRevisionResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) RenameRevision(
	ctx context.Context,
	request note.RenameRevisionRequestObject,
) (note.RenameRevisionResponseObject, error) {
	return nil, errors.New("not implemented")
}
