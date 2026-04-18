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
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, errs.NewInternalGenerateID(err)
	}
	var icon string
	if request.Body.Icon != nil {
		icon = *request.Body.Icon
	}

	cmd := &app.CreateNote{
		ID:       id,
		Name:     request.Body.Name,
		Icon:     icon,
		FolderID: *request.Body.FolderId,
		UserID:   user.ID,
	}
	err = h.App.Cmds.CreateNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.CreateNote201Response{
		Headers: note.CreateNote201ResponseHeaders{
			ContentLocation: h.BaseURL.JoinPath("notes", id.String()).String(),
		},
	}, nil
}

func (h *StrictHandler) PermanentlyDeleteNote(
	ctx context.Context,
	request note.PermanentlyDeleteNoteRequestObject,
) (note.PermanentlyDeleteNoteResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.PermanentlyDeleteNote{
		ID:     request.NoteId,
		UserID: user.ID,
	}
	err := h.App.Cmds.DeleteNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.PermanentlyDeleteNote204Response{}, nil
}

func (h *StrictHandler) GetNote(
	ctx context.Context,
	request note.GetNoteRequestObject,
) (note.GetNoteResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	excludeTrashed := request.Params.IncludeTrashed == nil || !*request.Params.IncludeTrashed

	query := &app.GetNote{
		ID:             request.NoteId,
		ExcludeTrashed: excludeTrashed,
		UserID:         user.ID,
	}

	result, err := h.App.Queries.GetNoteHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto, err := toNoteDTO(result)
	if err != nil {
		return nil, err
	}
	return note.GetNote200JSONResponse(dto), nil
}

func (h *StrictHandler) GetNoteGraph(
	ctx context.Context,
	request note.GetNoteGraphRequestObject,
) (note.GetNoteGraphResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	var depth int
	if request.Params.Depth != nil {
		depth = *request.Params.Depth
	}

	query := &app.GetNoteGraph{
		ID:     request.NoteId,
		Depth:  depth,
		UserID: user.ID,
	}

	result, err := h.App.Queries.GetNoteGraphHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toGraphDTO(&result)
	return note.GetNoteGraph200JSONResponse(dto), nil
}

func (h *StrictHandler) GetNoteLinks(
	ctx context.Context,
	request note.GetNoteLinksRequestObject,
) (note.GetNoteLinksResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}
	outgoingLinks := request.Params.OutgoingLinks != nil && *request.Params.OutgoingLinks
	backlinks := request.Params.Backlinks != nil && *request.Params.Backlinks
	query := &app.GetNoteLinks{
		ID:            request.NoteId,
		OutgoingLinks: outgoingLinks,
		Backlinks:     backlinks,
		UserID:        user.ID,
	}
	result, err := h.App.Queries.GetNoteLinksHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	dto := toGetNoteLinksDTO(result)
	return note.GetNoteLinks200JSONResponse(dto), nil
}

func (h *StrictHandler) PublishNote(
	ctx context.Context,
	request note.PublishNoteRequestObject,
) (note.PublishNoteResponseObject, error) {
	return nil, errs.Unimplemented
}

func (h *StrictHandler) RenameNote(
	ctx context.Context,
	request note.RenameNoteRequestObject,
) (note.RenameNoteResponseObject, error) {
	user, ok := commonhttp.UserFromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized
	}

	cmd := &app.RenameNote{
		ID:     request.NoteId,
		Name:   request.Body.Name,
		UserID: user.ID,
	}
	err := h.App.Cmds.RenameNoteHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return note.RenameNote204Response{}, nil
}

func (h *StrictHandler) UnpublishNote(
	ctx context.Context,
	request note.UnpublishNoteRequestObject,
) (note.UnpublishNoteResponseObject, error) {
	return nil, errs.Unimplemented
}
