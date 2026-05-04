package app

import (
	"context"

	"github.com/google/uuid"
)

type SearchSvc interface {
	GenerateWorkspaceToken(ctx context.Context, workspaceID uuid.UUID) (SearchToken, error)
}
