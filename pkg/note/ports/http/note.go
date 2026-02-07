package http

import (
	"context"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (s *strictServer) ListNotes(ctx context.Context, request note.ListNotesRequestObject) (note.ListNotesResponseObject, error) {
	return &note.ListNotes200JSONResponse{}, nil
}

func (s *strictServer) CreateNote(ctx context.Context, request note.CreateNoteRequestObject) (note.CreateNoteResponseObject, error) {
	return &note.CreateNote201JSONResponse{}, nil
}

func (s *strictServer) DeleteNote(ctx context.Context, request note.DeleteNoteRequestObject) (note.DeleteNoteResponseObject, error) {
	return &note.DeleteNote204Response{}, nil
}

func (s *strictServer) GetNote(ctx context.Context, request note.GetNoteRequestObject) (note.GetNoteResponseObject, error) {
	return &note.GetNote200JSONResponse{}, nil
}

func (s *strictServer) PatchNote(ctx context.Context, request note.PatchNoteRequestObject) (note.PatchNoteResponseObject, error) {
	return &note.PatchNote200JSONResponse{}, nil
}

func (s *strictServer) UpdateNote(ctx context.Context, request note.UpdateNoteRequestObject) (note.UpdateNoteResponseObject, error) {
	return &note.UpdateNote200JSONResponse{}, nil
}
