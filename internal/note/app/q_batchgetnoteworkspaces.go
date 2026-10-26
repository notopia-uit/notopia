package app

import (
	"context"

	"github.com/google/uuid"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type BatchGetNoteWorkspaces struct {
	NoteIDs []uuid.UUID
}

type BatchGetNoteWorkspacesHandler struct {
	readModel GetNotesWorkspaceIDsReadModel
}

func NewBatchGetNoteWorkspacesHandler(readModel GetNotesWorkspaceIDsReadModel) *BatchGetNoteWorkspacesHandler {
	return &BatchGetNoteWorkspacesHandler{readModel: readModel}
}

var ProvideBatchGetNoteWorkspacesHandler = NewBatchGetNoteWorkspacesHandler

type BatchGetNoteWorkspacesQuery commonhandler.Query[BatchGetNoteWorkspaces, map[uuid.UUID]uuid.UUID]

var _ BatchGetNoteWorkspacesQuery = (*BatchGetNoteWorkspacesHandler)(nil)

// Internal-only (gRPC from other services): no permission check.
func (h *BatchGetNoteWorkspacesHandler) Handle(ctx context.Context, q *BatchGetNoteWorkspaces) (map[uuid.UUID]uuid.UUID, error) {
	return h.readModel.Handle(ctx, q.NoteIDs)
}
