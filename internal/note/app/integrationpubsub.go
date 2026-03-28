package app

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
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

func (p *IntegrationPubSub) EventBus() *cqrs.EventBus {
	return p.eventBus
}

func (p *IntegrationPubSub) EventProcessor() *cqrs.EventProcessor {
	return p.eventProcessor
}

func (p *IntegrationPubSub) Router() *message.Router {
	return p.router
}

func (p *IntegrationPubSub) Run(ctx context.Context) error {
	return p.router.Run(ctx)
}

func (p *IntegrationPubSub) Close() error {
	return p.router.Close()
}
