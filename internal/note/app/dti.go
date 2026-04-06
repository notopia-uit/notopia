package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

// IF not need, remove this, if we don't need to holding slice or something, or we don't apply decorator
type DomainEventToIntegrationEventHandler[E domain.Event] interface {
	Handle(ctx context.Context, event E) error
}
