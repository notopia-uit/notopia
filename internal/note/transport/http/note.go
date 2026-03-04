package http

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) CreateNote(ctx context.Context, request note.CreateNoteRequestObject) (note.CreateNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GenerateDailyNote(
	ctx context.Context,
	request note.GenerateDailyNoteRequestObject,
) (note.GenerateDailyNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) SearchTags(
	ctx context.Context,
	request note.SearchTagsRequestObject,
) (note.SearchTagsResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) DeleteNote(
	ctx context.Context,
	request note.DeleteNoteRequestObject,
) (note.DeleteNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) GetNote(
	ctx context.Context,
	request note.GetNoteRequestObject,
) (note.GetNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) PublishNote(
	ctx context.Context,
	request note.PublishNoteRequestObject,
) (note.PublishNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) RenameNote(
	ctx context.Context,
	request note.RenameNoteRequestObject,
) (note.RenameNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) UnpublishNote(
	ctx context.Context,
	request note.UnpublishNoteRequestObject,
) (note.UnpublishNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}
