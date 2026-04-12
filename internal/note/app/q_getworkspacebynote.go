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
	GetWorkspaceByNote(ctx context.Context, noteID uuid.UUID) (*Workspace, error)
}

type GetWorkspaceByNoteHandler struct {
	authorizationService AuthorizationService
	readModel            GetWorkspaceByNoteReadModel
}

func NewGetWorkspaceByNoteHandler(
	authorizationService AuthorizationService,
	readModel GetWorkspaceByNoteReadModel,
) *GetWorkspaceByNoteHandler {
	return &GetWorkspaceByNoteHandler{
		authorizationService: authorizationService,
		readModel:            readModel,
	}
}

var ProvideGetWorkspaceByNoteHandler = NewGetWorkspaceByNoteHandler

func (h *GetWorkspaceByNoteHandler) Handle(ctx context.Context, query *GetWorkspaceByNote) (*Workspace, error) {
	workspace, err := h.readModel.GetWorkspaceByNote(ctx, query.NoteID)
	if err != nil {
		return nil, err
	}
	hasPermission, err := h.authorizationService.HasWorkspacePermission(ctx, query.UserID, workspace.ID, WorkspacePermissionRead)
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
