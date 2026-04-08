package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetNoteLinks struct {
	ID            uuid.UUID
	OutgoingLinks bool
	Backlinks     bool

	UserID string
}

type GetNoteLinksReadModel interface {
	GetNoteLinks(ctx context.Context, q *GetNoteLinks) (*NoteLinkResult, error)
}

type GetNoteLinksHandler struct {
	authorizationService AuthorizationService
	noteRepo             domain.NoteRepo
	readModel            GetNoteLinksReadModel
}

func NewGetNoteLinksHandler(
	authorizationService AuthorizationService,
	noteRepo domain.NoteRepo,
	readModel GetNoteLinksReadModel,
) *GetNoteLinksHandler {
	return &GetNoteLinksHandler{
		authorizationService: authorizationService,
		noteRepo:             noteRepo,
		readModel:            readModel,
	}
}

var ProvideGetNoteLinksHandler = NewGetNoteLinksHandler

func (h *GetNoteLinksHandler) Handle(ctx context.Context, query *GetNoteLinks) (*NoteLinkResult, error) {
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
		return nil, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read note links %s", query.UserID, query.ID),
		)
	}
	return h.readModel.GetNoteLinks(ctx, query)
}
