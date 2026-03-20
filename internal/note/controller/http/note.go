package http

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) GetNotes(
	ctx context.Context,
	request note.GetNotesRequestObject,
) (note.GetNotesResponseObject, error) {
	query := &app.GetNotes{
		ID: uuid.Nil,
	}
	result, err := h.App.GetNotesHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return note.GetNotes200JSONResponse{}, nil
	}

	dto := getNoteToDTO(result.Data[0])
	return note.GetNotes200JSONResponse(dto), nil
}

func (h *StrictHandler) CreateNote(
	ctx context.Context,
	request note.CreateNoteRequestObject,
) (note.CreateNoteResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("request body is required")
	}

	id := uuid.New()
	var folderId uuid.UUID
	if request.Body.FolderId != nil {
		folderId = *request.Body.FolderId
	}

	cmd := &app.CreateNote{
		ID:       id,
		Name:     request.Body.Name,
		Icon:     request.Body.Icon,
		Tags:     request.Body.Tags,
		FolderID: folderId,
	}
	err := h.App.CreateNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.CreateNote201Response{
		Headers: note.CreateNote201ResponseHeaders{
			ContentLocation: id.String(),
		},
	}, nil
}

func (h *StrictHandler) GenerateDailyNote(
	ctx context.Context,
	request note.GenerateDailyNoteRequestObject,
) (note.GenerateDailyNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}

func (h *StrictHandler) DeleteNote(
	ctx context.Context,
	request note.DeleteNoteRequestObject,
) (note.DeleteNoteResponseObject, error) {
	cmd := &app.DeleteNote{
		ID: request.NoteId,
	}
	err := h.App.DeleteNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.DeleteNote204Response{}, nil
}

func (h *StrictHandler) GetNote(
	ctx context.Context,
	request note.GetNoteRequestObject,
) (note.GetNoteResponseObject, error) {
	return nil, errors.New("not implemented - GetNoteHandler not available in app")
}

func (h *StrictHandler) GetNoteGraph(
	ctx context.Context,
	request note.GetNoteGraphRequestObject,
) (note.GetNoteGraphResponseObject, error) {
	query := &app.GetNoteGraph{
		ID:    request.NoteId,
		Depth: request.Params.Depth,
	}

	result, err := h.App.GetNoteGraphHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getGraphToDTO(*result)
	return note.GetNoteGraph200JSONResponse(dto), nil
}

func (h *StrictHandler) GetNoteLinks(
	ctx context.Context,
	request note.GetNoteLinksRequestObject,
) (note.GetNoteLinksResponseObject, error) {
	query := &app.GetNoteLinks{
		ID:            request.NoteId,
		OutgoingLinks: request.Params.OutgoingLinks,
		Backlinks:     request.Params.Backlinks,
	}
	result, err := h.App.GetNoteLinksHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getNoteLinkResultToDTO(*result)
	return note.GetNoteLinks200JSONResponse(dto), nil
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
	if request.Body == nil {
		return nil, errors.New("request body is required")
	}

	cmd := &app.RenameNote{
		ID:   request.NoteId,
		Name: request.Body.Name,
	}
	err := h.App.RenameNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.RenameNote204Response{}, nil
}

func (h *StrictHandler) UnpublishNote(
	ctx context.Context,
	request note.UnpublishNoteRequestObject,
) (note.UnpublishNoteResponseObject, error) {
	return nil, errors.New("not implemented")
}
