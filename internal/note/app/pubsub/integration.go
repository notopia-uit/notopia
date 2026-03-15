package pubsub

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Integration struct {
	eventBus       *cqrs.EventBus
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
	publisher      message.Publisher
}

func NewIntegration(
	eventBus *cqrs.EventBus,
	eventProcessor *cqrs.EventProcessor,
	router *message.Router,
	publisher message.Publisher,
) *Integration {
	return &Integration{
		eventBus:       eventBus,
		eventProcessor: eventProcessor,
		router:         router,
		publisher:      publisher,
	}
}

func (i *Integration) EventBus() *cqrs.EventBus {
	return i.eventBus
}

func (i *Integration) EventProcessor() *cqrs.EventProcessor {
	return i.eventProcessor
}

func (i *Integration) Router() *message.Router {
	return i.router
}

func (i *Integration) Publisher() message.Publisher {
	return i.publisher
}
