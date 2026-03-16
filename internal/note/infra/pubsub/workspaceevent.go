package pubsub

import (
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/redis/go-redis/v9"
)

// TODO: If have time, try https://github.com/stong1994/watermill-rediszset, because we only need pubsub, not stream
func NewWorkspaceEventInternalPubSub(
	logger watermill.LoggerAdapter,
	marshaler cqrs.CommandEventMarshaler,
	redisClient *RedisClient,
) (*pubsub.WorkspaceEventInternalPubSub, error) {
	topic := "events:workspaces"
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:        (*redis.Client)(redisClient),
		DefaultMaxlen: 10000,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis publisher: %w", err)
	}
	subcriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
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

	return pubsub.NewWorkspaceEventInternalPubSub(
		router,
		publisher,
		subcriber,
		topic,
	), nil
}

var ProvideWorkspaceEventInternalPubSub = NewWorkspaceEventInternalPubSub

func NewWorkspaceEventHubPubSub(
	logger watermill.LoggerAdapter,
) *pubsub.WorkspaceEventHubPubSub {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 100,
		},
		logger,
	)
	return pubsub.NewWorkspaceEventHubPubSub(
		pubSub,
	)
}

var ProvideWorkspaceEventHubPubSub = NewWorkspaceEventHubPubSub
