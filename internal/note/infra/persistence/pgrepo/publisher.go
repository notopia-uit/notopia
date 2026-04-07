package pgrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type Publisher interface {
	PublishWorkspaceItem(ctx context.Context, event domain.Event, workspaceID uuid.UUID) error
	Publish(ctx context.Context, event domain.Event) error
}

type PublisherFactory interface {
	Create(pgxTx pgx.Tx) (Publisher, error)
}
