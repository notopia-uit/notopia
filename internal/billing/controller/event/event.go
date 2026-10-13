package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
)

type Event struct {
	router *message.Router
	app    *app.Server
}

func New(
	kafkaCfg *commonconfig.Kafka,
	a *app.Server,
	tracer kafka.SaramaTracer,
	logger watermill.LoggerAdapter,
	_ otel.Global,
) (*Event, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create event controller router: %w", err)
	}
	retry := middleware.Retry{
		MaxRetries:      5,
		InitialInterval: time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2,
		Logger:          logger,
	}
	router.AddMiddleware(middleware.CorrelationID, middleware.Recoverer, retry.Middleware, wotel.Trace())

	e := &Event{router: router, app: a}
	handlers := []struct {
		name    string
		topic   string
		handler message.NoPublishHandlerFunc
	}{
		{name: "workspace-deleted", topic: component.TopicWorkspaceDeleted, handler: e.workspaceDeleted},
		{name: "role-updated", topic: component.TopicRoleUpdated, handler: e.roleUpdated},
		{name: "member-removed", topic: component.TopicMemberRemoved, handler: e.memberRemoved},
	}
	// One consumer group per topic: subscribers sharing a group steal each other's partitions.
	for _, h := range handlers {
		subscriber, err := kafka.NewSubscriber(
			kafka.SubscriberConfig{
				Brokers:               kafkaCfg.Brokers,
				ConsumerGroup:         kafkaCfg.ConsumerGroup + "." + h.name,
				Tracer:                tracer,
				OverwriteSaramaConfig: kafka.DefaultSaramaSubscriberConfig(),
			},
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s subscriber: %w", h.name, err)
		}
		router.AddConsumerHandler(h.name, h.topic, subscriber, h.handler)
	}
	return e, nil
}

var Provide = New

func (e *Event) workspaceDeleted(msg *message.Message) error {
	var ev share.WorkspaceDeletedIntegrationEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("failed to unmarshal WorkspaceDeletedIntegrationEvent: %w", err)
	}
	return e.app.Cmds.HandleWorkspaceDeleted.Handle(msg.Context(), &app.HandleWorkspaceDeleted{WorkspaceID: ev.Id})
}

func (e *Event) roleUpdated(msg *message.Message) error {
	var ev share.UserWorkspaceRoleUpdatedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("failed to unmarshal UserWorkspaceRoleUpdatedEvent: %w", err)
	}
	if ev.Role == share.Owner || ev.WorkspaceId == nil {
		return nil
	}
	return e.app.Cmds.HandlePayerRemoved.Handle(msg.Context(), &app.HandlePayerRemoved{
		WorkspaceID: *ev.WorkspaceId,
		UserID:      ev.UserId,
	})
}

func (e *Event) memberRemoved(msg *message.Message) error {
	var ev share.WorkspaceMemberRemovedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("failed to unmarshal WorkspaceMemberRemovedEvent: %w", err)
	}
	if ev.WorkspaceId == nil {
		return nil
	}
	return e.app.Cmds.HandlePayerRemoved.Handle(msg.Context(), &app.HandlePayerRemoved{
		WorkspaceID: *ev.WorkspaceId,
		UserID:      ev.UserId,
	})
}

func (e *Event) Run(ctx context.Context) error { return e.router.Run(ctx) }

func (e *Event) IsRunning() bool { return e.router.IsRunning() }
