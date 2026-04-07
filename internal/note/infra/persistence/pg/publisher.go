package pg

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type PublishWorkspaceItemParams struct {
	WorkspaceID uuid.UUID
	AggregateID uuid.UUID
}

type Publisher interface {
	PublishWorkspaceItem(ctx context.Context, event domain.Event, params PublishWorkspaceItemParams) error
	Publish(ctx context.Context, event domain.Event) error
}

type PublisherFactory interface {
	Create(pgxTx pgx.Tx) (Publisher, error)
}
