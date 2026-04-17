package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceByNote struct {
	NoteID uuid.UUID
	UserID string
}

type GetWorkspaceByNoteReadModel interface {
	GetWorkspaceByNoteID(ctx context.Context, noteID uuid.UUID) (*Workspace, error)
}

type GetWorkspaceByNoteHandler struct {
	authorizationSvc AuthorizationSvc
	readModel            GetWorkspaceByNoteReadModel
}

func NewGetWorkspaceByNoteHandler(
	authorizationSvc AuthorizationSvc,
	readModel GetWorkspaceByNoteReadModel,
) *GetWorkspaceByNoteHandler {
	return &GetWorkspaceByNoteHandler{
		authorizationSvc: authorizationSvc,
		readModel:            readModel,
	}
}

var ProvideGetWorkspaceByNoteHandler = NewGetWorkspaceByNoteHandler

func (h *GetWorkspaceByNoteHandler) Handle(ctx context.Context, query *GetWorkspaceByNote) (*Workspace, error) {
	workspace, err := h.readModel.GetWorkspaceByNoteID(ctx, query.NoteID)
	if err != nil {
		return nil, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(ctx, query.UserID, workspace.ID, WorkspacePermissionRead)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace %s", query.UserID, workspace.ID),
		)
	}
	return workspace, nil
}
