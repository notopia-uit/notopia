package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetWorkspaceByNote struct {
	NoteID uuid.UUID
	UserID string
}

type GetWorkspaceByNoteHandler struct {
	authorizationSvc AuthorizationSvc
	readModel        GetWorkspaceByNoteReadModel
}

func NewGetWorkspaceByNoteHandler(
	authorizationSvc AuthorizationSvc,
	readModel GetWorkspaceByNoteReadModel,
) *GetWorkspaceByNoteHandler {
	return &GetWorkspaceByNoteHandler{
		authorizationSvc: authorizationSvc,
		readModel:        readModel,
	}
}

var ProvideGetWorkspaceByNoteHandler = NewGetWorkspaceByNoteHandler

type GetWorkspaceByNoteQuery commonhandler.Query[GetWorkspaceByNote, Workspace]

var _ GetWorkspaceByNoteQuery = (*GetWorkspaceByNoteHandler)(nil)

func (h *GetWorkspaceByNoteHandler) Handle(ctx context.Context, query *GetWorkspaceByNote) (Workspace, error) {
	workspace, err := h.readModel.Handle(ctx, query.NoteID)
	if err != nil {
		return Workspace{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspacePermission(ctx, query.UserID, workspace.ID, WorkspacePermissionRead)
	if err != nil {
		return Workspace{}, err
	}
	if !hasPermission {
		return Workspace{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace %s", query.UserID, workspace.ID),
		)
	}
	return workspace, nil
}
