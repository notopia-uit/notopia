package app

import "github.com/notopia-uit/notopia/internal/note/domain"

type DomainEventToIntegrationEventHandler interface {
	Handle(event domain.Event) ([]IntegrationEvent, error)
}
