package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
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
		return nil, errs.NewInternalGenerateID(err)
	}

	cmd := &app.CreateNote{
		ID:       id,
		Name:     request.Body.Name,
		Icon:     request.Body.Icon,
		Tags:     request.Body.Tags,
		FolderID: *request.Body.FolderId,
		UserID:   user.ID,
	}
	err = h.App.CommandHandlers.CreateNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.CreateNote201Response{
		Headers: note.CreateNote201ResponseHeaders{
			ContentLocation: h.BaseURL.JoinPath("notes", id.String()).String(),
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

	cmd := &app.DeleteNote{
		ID:     request.NoteId,
		UserID: user.ID,
	}
	err = h.App.CommandHandlers.DeleteNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.DeleteNote204Response{}, nil
}

func (h *StrictHandler) GetNote(
	ctx context.Context,
	request note.GetNoteRequestObject,
) (note.GetNoteResponseObject, error) {
	user, err := commonhttp.UserFromContextError(ctx)
	if err != nil {
		return nil, err
	}

	query := &app.GetNote{
		ID:             request.NoteId,
		ExcludeTrashed: true,
		UserID:         user.ID,
	}

	result, err := h.App.QueryHandlers.GetNoteHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toNote(*result)
	return note.GetNote200JSONResponse(dto), nil
}

func (h *StrictHandler) GetNoteGraph(
	ctx context.Context,
	request note.GetNoteGraphRequestObject,
) (note.GetNoteGraphResponseObject, error) {
	var depth int
	if request.Params.Depth != nil {
		depth = *request.Params.Depth
	}

	query := &app.GetNoteGraph{
		ID:    request.NoteId,
		Depth: depth,
	}

	result, err := h.App.QueryHandlers.GetNoteGraphHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toGraph(result)
	return note.GetNoteGraph200JSONResponse(dto), nil
}

func (h *StrictHandler) GetNoteLinks(
	ctx context.Context,
	request note.GetNoteLinksRequestObject,
) (note.GetNoteLinksResponseObject, error) {
	outgoingLinks := request.Params.OutgoingLinks != nil && *request.Params.OutgoingLinks
	backlinks := request.Params.Backlinks != nil && *request.Params.Backlinks
	query := &app.GetNoteLinks{
		ID:            request.NoteId,
		OutgoingLinks: outgoingLinks,
		Backlinks:     backlinks,
	}
	result, err := h.App.QueryHandlers.GetNoteLinksHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toGetNoteLinks(result)
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

	cmd := &app.RenameNote{
		ID:     request.NoteId,
		Name:   request.Body.Name,
		UserID: user.ID,
	}
	err = h.App.CommandHandlers.RenameNoteHandler.Handle(ctx, cmd)
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
