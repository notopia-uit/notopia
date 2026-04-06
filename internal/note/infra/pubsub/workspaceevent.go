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
	"github.com/notopia-uit/notopia/internal/note/errs"
)

// NOTE: This should be refactor into redis event bus (holding redis publisher), publish to one single topic
// Then the redis subcriber ... idk, later
// The hub and the pubsub, so confused
const (
	MetadataWorkspaceIDKey = "workspaceId"
	metadataUserIDKey      = "userId"
	metadataEventTypeKey   = "eventType"
)

type WorkspaceEventInternalHub struct {
	router      *message.Router
	publisher   message.Publisher
	subscriber  message.Subscriber
	topic       string
	redisClient *RedisClient
}

// TODO: If have time, try https://github.com/stong1994/watermill-rediszset, because we only need pubsub, not stream
// This would reduce memory overhead and be more efficient for ephemeral workspace events.
func NewWorkspaceEventInternalHub(
	logger watermill.LoggerAdapter,
	marshaler *cqrs.JSONMarshaler,
	redisClient *RedisClient,
) (*WorkspaceEventInternalHub, error) {
	topic := "events:workspaces"
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:        redisClient,
		DefaultMaxlen: 10000,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis publisher: %w", err)
	}
	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:                        redisClient,
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

	return &WorkspaceEventInternalHub{
		router:      router,
		publisher:   publisher,
		subscriber:  subscriber,
		redisClient: redisClient,
		topic:       topic,
	}, nil
}

var ProvideWorkspaceEventInternalHub = NewWorkspaceEventInternalHub

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
	internalPubSub *WorkspaceEventInternalHub
	hubPubSub      *WorkspaceEventHubPubSub
}

var _ app.WorkspaceEventHub = (*WorkspaceEvent)(nil)

func NewWorkspaceEvent(
	internalPubSub *WorkspaceEventInternalHub,
	hubPubSub *WorkspaceEventHubPubSub,
) *WorkspaceEvent {
	internalPubSub.router.AddConsumerHandler(
		"handler",
		internalPubSub.topic,
		internalPubSub.subscriber,
		func(msg *message.Message) error {
			workspaceID := msg.Metadata.Get(MetadataWorkspaceIDKey)
			return hubPubSub.pubSub.Publish(workspaceID, msg.Copy())
		},
	)
	return &WorkspaceEvent{
		internalPubSub: internalPubSub,
		hubPubSub:      hubPubSub,
	}
}

var ProvideWorkspaceEvent = NewWorkspaceEvent

func (w *WorkspaceEvent) Publish(ctx context.Context, workspaceID uuid.UUID, userID string, events ...app.WorkspaceEvent) error {
	msgs := make([]*message.Message, 0, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return errs.NewWorkspaceEventPubSubFailedToCreateMessage(
				userID,
				workspaceID,
				err,
			)
		}
		msg := message.NewMessage(watermill.NewUUID(), payload)
		msg.Metadata.Set(MetadataWorkspaceIDKey, fmt.Sprintf("%v", workspaceID))
		msg.Metadata.Set(metadataUserIDKey, userID)
		msg.Metadata.Set(metadataEventTypeKey, event.GetEvent())
		msgs = append(msgs, msg)
	}
	err := w.internalPubSub.publisher.Publish(w.internalPubSub.topic, msgs...)
	if err != nil {
		return errs.NewWorkspaceEventPubSubPublishFailed(
			userID,
			workspaceID,
			err,
		)
	}
	return nil
}

func (w *WorkspaceEvent) Subscribe(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) (<-chan app.WorkspaceEvent, error) {
	eventCh := make(chan app.WorkspaceEvent, 10)

	msgCh, err := w.hubPubSub.pubSub.Subscribe(ctx, fmt.Sprintf("%v", workspaceID))
	if err != nil {
		return nil, errs.NewWorkspaceEventPubSubSubscribeFailed(
			userID,
			workspaceID,
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
				if msg.Metadata.Get(metadataUserIDKey) == userID {
					msg.Ack()
					continue
				}
				eventType := msg.Metadata.Get(metadataEventTypeKey)
				if eventType == "" {
					slog.ErrorContext(
						ctx, "missing event type in message metadata",
						slog.String("workspace_id", workspaceID.String()),
						slog.String("user_id", userID),
					)
					msg.Ack()
					continue
				}
				event, ok := app.NewEmptyWorkspaceEventFromType(eventType)
				if !ok {
					slog.ErrorContext(
						ctx, "unknown event type in message metadata",
						slog.String("event_type", eventType),
						slog.String("workspace_id", workspaceID.String()),
						slog.String("user_id", userID),
					)
					msg.Ack()
					continue
				}
				if err := json.Unmarshal(msg.Payload, event); err != nil {
					slog.ErrorContext(ctx, "failed to unmarshal event", slog.Any("error", err))
					msg.Ack()
					continue
				}
				select {
				case eventCh <- event:
					msg.Ack()
				case <-ctx.Done():
					return
				default:
					slog.WarnContext(ctx, "dropping event", slog.String("workspace_id", workspaceID.String()), slog.String("user_id", userID))
					msg.Ack()
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

func (w *WorkspaceEvent) Check(ctx context.Context) error {
	if statusCmd := w.internalPubSub.redisClient.Ping(ctx); statusCmd.Err() != nil {
		return fmt.Errorf("failed to ping Redis: %w", statusCmd.Err())
	}
	return nil
}
