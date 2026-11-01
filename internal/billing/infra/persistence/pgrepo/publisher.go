package pgrepo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/billing/domain"
)

// EventPublisher writes domain events into the transaction it was created for.
type EventPublisher interface {
	Publish(ctx context.Context, events ...domain.Event) error
}

type PublisherFactory interface {
	Create(tx pgx.Tx) (EventPublisher, error)
}
