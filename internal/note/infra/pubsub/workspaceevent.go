package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
	"github.com/redis/go-redis/v9"
)

const (
	MetadataWorkspaceIDKey = "workspace_id"
	MetadataUserIDKey      = "user_id"
	MetadataEventTypeKey   = "event_type"
)

type WorkspaceEventInternalPubSub struct {
	router     *message.Router
	publisher  message.Publisher
	subscriber message.Subscriber
	topic      string
}

// TODO: If have time, try https://github.com/stong1994/watermill-rediszset, because we only need pubsub, not stream
func NewWorkspaceEventInternalPubSub(
	logger watermill.LoggerAdapter,
	marshaler cqrs.CommandEventMarshaler,
	redisClient *RedisClient,
) (*WorkspaceEventInternalPubSub, error) {
	topic := "events:workspaces"
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:        (*redis.Client)(redisClient),
		DefaultMaxlen: 10000,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis publisher: %w", err)
	}
	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:                        (*redis.Client)(redisClient),
		FanOutOldestId:                "$",
		DisableIndefiniteInitialBlock: true,
		BlockTime:                     2 * time.Second,
	}, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis subscriber: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal message router: %w", err)
	}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer)

	return &WorkspaceEventInternalPubSub{
		router:     router,
		publisher:  publisher,
		subscriber: subscriber,
		topic:      topic,
	}, nil
}

var ProvideWorkspaceEventInternalPubSub = NewWorkspaceEventInternalPubSub

type WorkspaceEventHubPubSub struct {
	pubSub *gochannel.GoChannel
}

func NewWorkspaceEventHubPubSub(
	logger watermill.LoggerAdapter,
) *WorkspaceEventHubPubSub {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 100,
		},
		logger,
	)
	return &WorkspaceEventHubPubSub{
		pubSub: pubSub,
	}
}

var ProvideWorkspaceEventHubPubSub = NewWorkspaceEventHubPubSub

type WorkspaceEvent struct {
	internalPubSub *WorkspaceEventInternalPubSub
	hubPubSub      *WorkspaceEventHubPubSub
}

var _ app.WorkspaceEventPubSub = (*WorkspaceEvent)(nil)

func NewWorkspaceEvent(
	internalPubSub *WorkspaceEventInternalPubSub,
	hubPubSub *WorkspaceEventHubPubSub,
) *WorkspaceEvent {
	internalPubSub.router.AddConsumerHandler(
		"handler",
		internalPubSub.topic,
		internalPubSub.subscriber,
		func(msg *message.Message) error {
			workspaceID := msg.Metadata.Get(MetadataWorkspaceIDKey)
			return hubPubSub.pubSub.Publish(workspaceID, msg)
		},
	)
	return &WorkspaceEvent{
		internalPubSub: internalPubSub,
		hubPubSub:      hubPubSub,
	}
}

var ProvideWorkspaceEvent = NewWorkspaceEvent

func (w *WorkspaceEvent) Publish(ctx context.Context, workspaceID uuid.UUID, userID string, events ...domain.Event) error {
	msgs := make([]*message.Message, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		msg.Metadata.Set(MetadataWorkspaceIDKey, fmt.Sprintf("%v", workspaceID))
		msg.Metadata.Set(MetadataUserIDKey, userID)
		msg.Metadata.Set(MetadataEventTypeKey, string(event.EventType()))
		msg.SetContext(ctx)
		msgs = append(msgs, msg)
	}
	err := w.internalPubSub.publisher.Publish(w.internalPubSub.topic, msgs...)
	if err != nil {
		return commonerror.NewInternal(
			fmt.Sprintf("Failed to publish workspace events for workspace %q and user %q", workspaceID.String(), userID),
			"PublishWorkspaceEventsFailed",
			err,
		)
	}
	return nil
}

func (w *WorkspaceEvent) Subscribe(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) (<-chan domain.Event, error) {
	eventCh := make(chan domain.Event, 10)

	msgCh, err := w.hubPubSub.pubSub.Subscribe(ctx, fmt.Sprintf("%v", workspaceID))
	if err != nil {
		return nil, commonerror.NewInternal(
			fmt.Sprintf("Failed to subscribe to workspace events for workspace %q and user %q", workspaceID.String(), userID),
			"SubscribeWorkspaceEventsFailed",
			err,
		)
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
	return w.internalPubSub.router.Run(ctx)
}

func (w *WorkspaceEvent) Close() error {
	var errs []error

	if err := w.internalPubSub.router.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := w.internalPubSub.publisher.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := w.internalPubSub.subscriber.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := w.hubPubSub.pubSub.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(append([]error{fmt.Errorf("failed to close workspace event pubsub")}, errs...)...)
	}
	return nil
}
