package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

type GetNote struct {
	ID uuid.UUID

	UserID string
}

type GetNoteReadModel interface {
	GetNote(ctx context.Context, q *GetNote) (*Note, error)
}

type GetNoteHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	readModel            GetNoteReadModel
}

func NewGetNoteHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	readModel GetNoteReadModel,
) *GetNoteHandler {
	return &GetNoteHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		readModel:            readModel,
	}
}

var ProvideGetNoteHandler = NewGetNoteHandler

func (h *GetNoteHandler) Handle(ctx context.Context, query *GetNote) (*Note, error) {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	hasPermission, err := h.authorizationService.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		workspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, newErrGetNoteForbidden(query.UserID, workspaceID)
	}
	return h.readModel.GetNote(ctx, query)
}

var ErrCodeGetNoteForbidden = "GetNote_1"

func newErrGetNoteForbidden(userID string, workspaceID uuid.UUID) *commonerror.Err {
	return commonerror.NewForbidden(
		fmt.Sprintf("user %q does not have permission to access workspace %q", userID, workspaceID),
		ErrCodeGetNoteForbidden,
		nil,
	)
}
