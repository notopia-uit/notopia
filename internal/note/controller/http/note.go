package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app/command"
	"github.com/notopia-uit/notopia/internal/note/app/query"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

func (h *StrictHandler) CreateNote(
	ctx context.Context,
	request note.CreateNoteRequestObject,
) (note.CreateNoteResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, errs.NewInternal("failed to generate UUIDv7 for new note", err)
	}

	cmd := &command.CreateNote{
		ID:       id,
		Name:     request.Body.Name,
		Icon:     request.Body.Icon,
		Tags:     request.Body.Tags,
		FolderID: *request.Body.FolderId,
		UserID:   user.ID,
	}
	err = h.App.CreateNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.CreateNote201Response{
		Headers: note.CreateNote201ResponseHeaders{
			ContentLocation: h.ServerURL + "/notes/" + id.String(),
		},
	}, nil
}

func (h *StrictHandler) GenerateDailyNote(
	ctx context.Context,
	request note.GenerateDailyNoteRequestObject,
) (note.GenerateDailyNoteResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) DeleteNote(
	ctx context.Context,
	request note.DeleteNoteRequestObject,
) (note.DeleteNoteResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &command.DeleteNote{
		ID:     request.NoteId,
		UserID: user.ID,
	}
	err = h.App.DeleteNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.DeleteNote204Response{}, nil
}

func (h *StrictHandler) GetNote(
	ctx context.Context,
	request note.GetNoteRequestObject,
) (note.GetNoteResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) GetNoteGraph(
	ctx context.Context,
	request note.GetNoteGraphRequestObject,
) (note.GetNoteGraphResponseObject, error) {
	var depth int
	if request.Params.Depth != nil {
		depth = *request.Params.Depth
	}

	query := &query.GetNoteGraph{
		ID:    request.NoteId,
		Depth: depth,
	}

	result, err := h.App.GetNoteGraphHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getGraphToDTO(result)
	return note.GetNoteGraph200JSONResponse(dto), nil
}

func (h *StrictHandler) GetNoteLinks(
	ctx context.Context,
	request note.GetNoteLinksRequestObject,
) (note.GetNoteLinksResponseObject, error) {
	outgoingLinks := request.Params.OutgoingLinks != nil && *request.Params.OutgoingLinks
	backlinks := request.Params.Backlinks != nil && *request.Params.Backlinks
	query := &query.GetNoteLinks{
		ID:            request.NoteId,
		OutgoingLinks: outgoingLinks,
		Backlinks:     backlinks,
	}
	result, err := h.App.GetNoteLinksHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := getNoteLinkResultToDTO(result)
	return note.GetNoteLinks200JSONResponse(dto), nil
}

func (h *StrictHandler) PublishNote(
	ctx context.Context,
	request note.PublishNoteRequestObject,
) (note.PublishNoteResponseObject, error) {
	return nil, errs.NewUnimplemented()
}

func (h *StrictHandler) RenameNote(
	ctx context.Context,
	request note.RenameNoteRequestObject,
) (note.RenameNoteResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &command.RenameNote{
		ID:     request.NoteId,
		Name:   request.Body.Name,
		UserID: user.ID,
	}
	err = h.App.RenameNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.RenameNote204Response{}, nil
}

func (h *StrictHandler) UnpublishNote(
	ctx context.Context,
	request note.UnpublishNoteRequestObject,
) (note.UnpublishNoteResponseObject, error) {
	return nil, errs.NewUnimplemented()
}
