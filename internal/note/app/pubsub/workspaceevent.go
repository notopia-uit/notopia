package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

const (
	MetadataWorkspaceIDKey = "workspace_id"
	MetadataUserIDKey      = "user_id"
	MetadataEventTypeKey   = "event_type"
)

type WorkspaceEventInternalPubSub struct {
	Router    *message.Router
	Publisher message.Publisher
	Subcriber message.Subscriber
	Topic     string
}

func NewWorkspaceEventInternalPubSub(
	router *message.Router,
	publisher message.Publisher,
	subscriber message.Subscriber,
	topic string,
) *WorkspaceEventInternalPubSub {
	return &WorkspaceEventInternalPubSub{
		Router:    router,
		Publisher: publisher,
		Subcriber: subscriber,
		Topic:     topic,
	}
}

type WorkspaceEventHubPubSub struct {
	PubSub *gochannel.GoChannel
}

func NewWorkspaceEventHubPubSub(
	pubSub *gochannel.GoChannel,
) *WorkspaceEventHubPubSub {
	return &WorkspaceEventHubPubSub{
		PubSub: pubSub,
	}
}

type WorkspaceEvent struct {
	InternalPubSub *WorkspaceEventInternalPubSub
	HubPubSub      *WorkspaceEventHubPubSub
}

func (w *WorkspaceEvent) Setup() {
	w.InternalPubSub.Router.AddConsumerHandler(
		"handler",
		w.InternalPubSub.Topic,
		w.InternalPubSub.Subcriber,
		func(msg *message.Message) error {
			workspaceID := msg.Metadata.Get(MetadataWorkspaceIDKey)
			return w.HubPubSub.PubSub.Publish(workspaceID, msg)
		},
	)
}

func (w *WorkspaceEvent) Publish(ctx context.Context, workspaceID uuid.UUID, userID string, event domain.WorkspaceEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set(MetadataWorkspaceIDKey, workspaceID.String())
	msg.Metadata.Set(MetadataUserIDKey, userID)
	msg.Metadata.Set(MetadataEventTypeKey, string(event.EventType()))
	msg.SetContext(ctx)
	return w.InternalPubSub.Publisher.Publish(w.InternalPubSub.Topic, msg)
}

func (w *WorkspaceEvent) Subscribe(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) (<-chan domain.WorkspaceEvent, error) {
	eventCh := make(chan domain.WorkspaceEvent, 10)

	msgCh, err := w.HubPubSub.PubSub.Subscribe(ctx, workspaceID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to workspace events: %w", err)
	}

	go func() {
		defer close(eventCh)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				if msg.Metadata.Get(MetadataUserIDKey) != userID {
					msg.Ack()
					continue
				}
				eventType := msg.Metadata.Get(MetadataEventTypeKey)
				if eventType == "" {
					slog.ErrorContext(ctx, "missing event type in message metadata", slog.String("workspace_id", workspaceID.String()), slog.String("user_id", userID))
					msg.Ack()
					continue
				}
				event, ok := domain.NewFromEventType(eventType)
				if !ok {
					slog.ErrorContext(ctx, "unknown event type in message metadata", slog.String("event_type", eventType), slog.String("workspace_id", workspaceID.String()), slog.String("user_id", userID))
					msg.Ack()
					continue
				}
				if err := json.Unmarshal(msg.Payload, event); err != nil {
					slog.ErrorContext(ctx, "failed to unmarshal event", slog.String("error", err.Error()))
					msg.Ack()
					continue
				}
				select {
				case eventCh <- event:
					msg.Ack()
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return eventCh, nil
}

func (w *WorkspaceEvent) Run(ctx context.Context) error {
	return w.InternalPubSub.Router.Run(ctx)
}

func (w *WorkspaceEvent) Close() error {
	var errs []error

	if err := w.InternalPubSub.Router.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := w.InternalPubSub.Publisher.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := w.InternalPubSub.Subcriber.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := w.HubPubSub.PubSub.Close(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(fmt.Errorf("failed to close workspace event pubsub"), errors.Join(errs...))
}
