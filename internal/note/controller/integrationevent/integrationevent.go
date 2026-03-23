package integrationevent

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
)

type IntegrationPubSub struct {
	eventBus       *cqrs.EventBus
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
}

func NewIntegrationPubSub(
	eventBus *cqrs.EventBus,
	eventProcessor *cqrs.EventProcessor,
	router *message.Router,
) *IntegrationPubSub {
	return &IntegrationPubSub{
		eventBus:       eventBus,
		eventProcessor: eventProcessor,
		router:         router,
	}
}

type IntegrationEvent struct {
	IntegrationPubSub
	app *app.App
}

func NewIntegrationEvent(
	integrationPubSub *IntegrationPubSub,
	app *app.App,
) (*IntegrationEvent, error) {
	err := integrationPubSub.eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"DocumentCommittedHandler",
			app.DocumentCommittedHandler.Handle,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add event handlers to integration event processor: %w", err)
	}

	return &IntegrationEvent{
		IntegrationPubSub: *integrationPubSub,
		app:               app,
	}, nil
}

var ProvideIntegrationEvent = NewIntegrationEvent

func (i *IntegrationEvent) Run(ctx context.Context) error {
	return i.router.Run(ctx)
}

func (i *IntegrationEvent) Close() error {
	return i.router.Close()
}
