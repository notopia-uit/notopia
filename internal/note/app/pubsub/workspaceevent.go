package pubsub

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

const MetadataWorkspaceIDKey = "workspace_id"

// NOTE: If need to provide these separately, need to create type definition for each
type WorkspaceEvent struct {
	eventBus       *cqrs.EventBus
	eventProcessor *cqrs.EventProcessor
	router         *message.Router
	publisher      message.Publisher
	subcriber      message.Subscriber
	topic          string
}

func NewWorkspaceEvent(
	eventBus *cqrs.EventBus,
	eventProcessor *cqrs.EventProcessor,
	router *message.Router,
	publisher message.Publisher,
	subscriber message.Subscriber,
	topic string,
) *WorkspaceEvent {
	return &WorkspaceEvent{
		eventBus:       eventBus,
		eventProcessor: eventProcessor,
		router:         router,
		publisher:      publisher,
		subcriber:      subscriber,
		topic:          topic,
	}
}

func (w *WorkspaceEvent) EventBus() *cqrs.EventBus {
	return w.eventBus
}

func (w *WorkspaceEvent) EventProcessor() *cqrs.EventProcessor {
	return w.eventProcessor
}

func (w *WorkspaceEvent) Router() *message.Router {
	return w.router
}

func (w *WorkspaceEvent) Publisher() message.Publisher {
	return w.publisher
}

func (w *WorkspaceEvent) Subscriber() message.Subscriber {
	return w.subcriber
}

func (w *WorkspaceEvent) Subcribe(ctx context.Context) (<-chan *message.Message, error) {
	return w.subcriber.Subscribe(ctx, w.topic)
}
