package server

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHTTPHandler) ListNotes(ctx context.Context, request note.ListNotesRequestObject) (note.ListNotesResponseObject, error) {
	return &note.ListNotes200JSONResponse{}, nil
}

func (h *StrictHTTPHandler) CreateNote(ctx context.Context, request note.CreateNoteRequestObject) (note.CreateNoteResponseObject, error) {
	return &note.CreateNote201JSONResponse{}, nil
}

func (h *StrictHTTPHandler) DeleteNote(ctx context.Context, request note.DeleteNoteRequestObject) (note.DeleteNoteResponseObject, error) {
	return &note.DeleteNote204Response{}, nil
}

func (h *StrictHTTPHandler) GetNote(ctx context.Context, request note.GetNoteRequestObject) (note.GetNoteResponseObject, error) {
	return &note.GetNote200JSONResponse{}, nil
}

func (h *StrictHTTPHandler) PatchNote(ctx context.Context, request note.PatchNoteRequestObject) (note.PatchNoteResponseObject, error) {
	return &note.PatchNote200JSONResponse{}, nil
}

func (h *StrictHTTPHandler) UpdateNote(ctx context.Context, request note.UpdateNoteRequestObject) (note.UpdateNoteResponseObject, error) {
	return &note.UpdateNote200JSONResponse{}, nil
}
