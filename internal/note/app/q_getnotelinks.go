package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetNoteLinks struct {
	ID            uuid.UUID
	OutgoingLinks bool
	Backlinks     bool

	UserID string
}

type GetNoteLinksHandler struct {
	authorizationSvc AuthorizationSvc
	noteRepo         domain.NoteRepo
	readModel        GetNoteLinksReadModel
}

func NewGetNoteLinksHandler(
	authorizationSvc AuthorizationSvc,
	noteRepo domain.NoteRepo,
	readModel GetNoteLinksReadModel,
) *GetNoteLinksHandler {
	return &GetNoteLinksHandler{
		authorizationSvc: authorizationSvc,
		noteRepo:         noteRepo,
		readModel:        readModel,
	}
}

var ProvideGetNoteLinksHandler = NewGetNoteLinksHandler

type GetNoteLinksQuery commonhandler.Query[GetNoteLinks, NoteLinkResult]

var _ GetNoteLinksQuery = (*GetNoteLinksHandler)(nil)

func (h *GetNoteLinksHandler) Handle(ctx context.Context, query *GetNoteLinks) (NoteLinkResult, error) {
	workspaceID, err := h.noteRepo.GetWorkspaceIDByID(ctx, query.ID)
	if err != nil {
		return NoteLinkResult{}, err
	}
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		workspaceID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return NoteLinkResult{}, err
	}
	if !hasPermission {
		return NoteLinkResult{}, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read note links %s", query.UserID, query.ID),
		)
	}
	result, err := h.readModel.Handle(ctx, &GetNoteLinksReadModelParams{
		ID:            query.ID,
		OutgoingLinks: query.OutgoingLinks,
		Backlinks:     query.Backlinks,
	})
	if err != nil {
		return NoteLinkResult{}, err
	}
	return result, nil
}
