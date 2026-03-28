package integrationevent

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/notopia-uit/notopia/internal/note/app"
)

type IntegrationEvent struct {
	pubSub *app.IntegrationPubSub
	app    *app.Server
}

func NewIntegrationEvent(
	integrationPubSub *app.IntegrationPubSub,
	app *app.Server,
) (*IntegrationEvent, error) {
	err := integrationPubSub.EventProcessor().AddHandlers(
		cqrs.NewEventHandler(
			"DocumentCommittedHandler",
			app.IntegrationEventHandlers.DocumentCommittedHandler.Handle,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add event handlers to integration event processor: %w", err)
	}

	return &IntegrationEvent{
		pubSub: integrationPubSub,
		app:    app,
	}, nil
}

var ProvideIntegrationEvent = NewIntegrationEvent
