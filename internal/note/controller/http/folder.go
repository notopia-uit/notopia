package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) CreateFolder(
	ctx context.Context,
	request note.CreateFolderRequestObject,
) (note.CreateFolderResponseObject, error) {
	body := request.Body
	if body == nil {
		return note.CreateFolder400JSONResponse{BadRequestErrorJSONResponse: note.BadRequestErrorJSONResponse{Message: "request body required"}}, nil
	}

	if body.ParentId == nil {
		return note.CreateFolder400JSONResponse{BadRequestErrorJSONResponse: note.BadRequestErrorJSONResponse{Message: "parentId is required"}}, nil
	}
	if body.WorkspaceId == nil {
		return note.CreateFolder400JSONResponse{BadRequestErrorJSONResponse: note.BadRequestErrorJSONResponse{Message: "workspaceId is required"}}, nil
	}

	id := uuid.New()
	cmd := &app.CreateFolder{
		ID:          id,
		Name:        body.Name,
		Icon:        body.Icon,
		ParentID:    uuid.UUID(*body.ParentId),
		WorkspaceID: uuid.UUID(*body.WorkspaceId),
	}

	if err := h.App.CreateFolderHandler.Handle(ctx, cmd); err != nil {
		return mapCreateFolderErrorResponse(err), nil
	}

	return note.CreateFolder201Response{
		Headers: note.CreateFolder201ResponseHeaders{
			ContentLocation: id.String(),
		},
	}, nil
}

func (h *StrictHandler) RenameFolder(
	ctx context.Context,
	request note.RenameFolderRequestObject,
) (note.RenameFolderResponseObject, error) {
	body := request.Body
	if body == nil {
		return note.RenameFolder400JSONResponse{BadRequestErrorJSONResponse: note.BadRequestErrorJSONResponse{Message: "request body required"}}, nil
	}

	cmd := &app.RenameFolder{
		ID:   uuid.UUID(request.FolderId),
		Name: body.Name,
	}

	if err := h.App.RenameFolderHandler.Handle(ctx, cmd); err != nil {
		return mapRenameFolderErrorResponse(err), nil
	}

	return note.RenameFolder204Response{}, nil
}

func mapCreateFolderErrorResponse(err error) note.CreateFolderResponseObject {
	return note.CreateFolder400JSONResponse{
		BadRequestErrorJSONResponse: note.BadRequestErrorJSONResponse{
			Message: err.Error(),
		},
	}
}

func mapRenameFolderErrorResponse(err error) note.RenameFolderResponseObject {
	return note.RenameFolder400JSONResponse{
		BadRequestErrorJSONResponse: note.BadRequestErrorJSONResponse{
			Message: err.Error(),
		},
	}
}
