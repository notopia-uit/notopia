package app

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

// This has to be instantiate each time of handling, under persistence transaction
// Mainly used by the unit of work
type DomainEventPublisher interface {
	Publish(ctx context.Context, events ...domain.Event) error
}

type DomainEventHandler[E domain.Event] interface {
	Handle(ctx context.Context, event E) error
}

type DomainEventProcessor interface {
	RegisterHandlers(handlers ...DomainEventHandler[domain.Event]) error
	Run(ctx context.Context) error
	Close() error
}
