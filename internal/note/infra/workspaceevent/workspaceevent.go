package workspaceevent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type WorkspaceEventHub struct {
	redisPublisher  message.Publisher
	redisSubscriber message.Subscriber
	internalPubSub  *gochannel.GoChannel
	router          *message.Router

	topic                  string
	MetadataWorkspaceIDKey string
	metadataUserIDKey      string
	metadataEventTypeKey   string
}

var _ app.WorkspaceEventHub = (*WorkspaceEventHub)(nil)

func NewWorkspaceEventHub(
	workspaceEventCfg *config.WorkspaceEvent,
	logger watermill.LoggerAdapter,
	redisClient *RedisClient,
) (*WorkspaceEventHub, error) {
	// TODO: If have time, try https://github.com/stong1994/watermill-rediszset, because we only need pubsub, not stream
	// This would reduce memory overhead and be more efficient for ephemeral workspace events.
	// Single Redis subscriber for ALL workspace events
	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client:        redisClient,
			DefaultMaxlen: 10000,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace event hub redis publisher: %w", err)
	}

	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:                        redisClient,
			FanOutOldestId:                "$",
			DisableIndefiniteInitialBlock: true,
			BlockTime:                     2 * time.Second,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace event hub redis subscriber: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace event hub router: %w", err)
	}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer)

	return &WorkspaceEventHub{
		redisPublisher:  publisher,
		redisSubscriber: subscriber,
		internalPubSub:  gochannel.NewGoChannel(gochannel.Config{OutputChannelBuffer: 100}, logger),
		router:          router,

		topic:                  workspaceEventCfg.MessageGeneralTopic,
		metadataUserIDKey:      workspaceEventCfg.MessageMetadataUserIDKey,
		metadataEventTypeKey:   workspaceEventCfg.MessageMetadataEventTypeKey,
		MetadataWorkspaceIDKey: workspaceEventCfg.MessageMetadataWorkspaceIDKey,
	}, nil
}

var ProvideWorkspaceEventHub = NewWorkspaceEventHub

func (h *WorkspaceEventHub) setupRouting() {
	h.router.AddHandler(
		"workspace-router",
		h.topic,
		h.redisSubscriber,
		"",
		h.internalPubSub,
		func(msg *message.Message) ([]*message.Message, error) {
			workspaceID := msg.Metadata.Get(h.MetadataWorkspaceIDKey)
			if workspaceID == "" {
				slog.ErrorContext(msg.Context(), "missing workspace ID in message metadata")
				return []*message.Message{}, nil
			}

			internalTopic := fmt.Sprintf("workspace:%s", workspaceID)
			if err := h.internalPubSub.Publish(internalTopic, msg.Copy()); err != nil {
				slog.ErrorContext(msg.Context(), "failed to publish to internal topic", slog.Any("error", err))
			}

			return []*message.Message{}, nil
		},
	)
}

func (h *WorkspaceEventHub) Publish(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
	events ...app.WorkspaceEvent,
) error {
	if len(events) == 0 {
		return nil
	}

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
		msg.SetContext(ctx)
		msg.Metadata.Set(h.MetadataWorkspaceIDKey, workspaceID.String())
		msg.Metadata.Set(h.metadataUserIDKey, userID)
		msg.Metadata.Set(h.metadataEventTypeKey, event.GetEvent())
		msgs = append(msgs, msg)
	}

	if err := h.redisPublisher.Publish(h.topic, msgs...); err != nil {
		return errs.NewWorkspaceEventPubSubPublishFailed(
			userID,
			workspaceID,
			err,
		)
	}

	return nil
}

func (h *WorkspaceEventHub) Subscribe(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) (<-chan app.WorkspaceEvent, error) {
	eventCh := make(chan app.WorkspaceEvent, 10)
	workspaceKey := fmt.Sprintf("workspace:%s", workspaceID)

	msgCh, err := h.internalPubSub.Subscribe(ctx, workspaceKey)
	if err != nil {
		return nil, errs.NewWorkspaceEventPubSubSubscribeFailed(userID, workspaceID, err)
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
				h.processMessage(ctx, msg, eventCh, userID, workspaceID)
			}
		}
	}()

	return eventCh, nil
}

func (h *WorkspaceEventHub) processMessage(
	ctx context.Context,
	msg *message.Message,
	eventCh chan app.WorkspaceEvent,
	userID string,
	workspaceID uuid.UUID,
) {
	// Skip own events
	if msg.Metadata.Get(h.metadataUserIDKey) == userID {
		msg.Ack()
		return
	}

	eventType := msg.Metadata.Get(h.metadataEventTypeKey)
	if eventType == "" {
		slog.ErrorContext(ctx, "missing event type in message metadata",
			slog.String("workspace_id", workspaceID.String()),
			slog.String("user_id", userID),
		)
		msg.Ack()
		return
	}

	event, ok := app.NewEmptyWorkspaceEventFromType(eventType)
	if !ok {
		slog.ErrorContext(ctx, "unknown event type",
			slog.String("event_type", eventType),
			slog.String("workspace_id", workspaceID.String()),
			slog.String("user_id", userID),
		)
		msg.Ack()
		return
	}

	if err := json.Unmarshal(msg.Payload, event); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal event",
			slog.Any("error", err),
			slog.String("workspace_id", workspaceID.String()),
		)
		msg.Ack()
		return
	}

	select {
	case eventCh <- event:
		msg.Ack()
	case <-ctx.Done():
		return
	default:
		slog.WarnContext(ctx, "dropping event (subscriber buffer full)",
			slog.String("workspace_id", workspaceID.String()),
			slog.String("user_id", userID),
			slog.String("event_type", eventType),
		)
		msg.Ack()
	}
}

func (h *WorkspaceEventHub) Run(ctx context.Context) error {
	h.setupRouting()
	return h.router.Run(ctx)
}

func (h *WorkspaceEventHub) Close() error {
	var errs []error

	if err := h.router.Close(); err != nil {
		errs = append(errs, fmt.Errorf("router close: %w", err))
	}

	if err := h.redisPublisher.Close(); err != nil {
		errs = append(errs, fmt.Errorf("publisher close: %w", err))
	}

	// NOTE: actually we don't need to close the Redis subscriber and the go pubsub because router already closed them
	if err := h.redisSubscriber.Close(); err != nil {
		errs = append(errs, fmt.Errorf("subscriber close: %w", err))
	}

	if err := h.internalPubSub.Close(); err != nil {
		errs = append(errs, fmt.Errorf("internal pubsub close: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close workspace event hub: %v", errors.Join(errs...))
	}
	return nil
}
